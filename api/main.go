package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/apex/gateway"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	scalev1 "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	shared "github.com/iLert/kubernetes-alerting-lambda-sample"
)

func main() {
	addr := fmt.Sprintf(":%s", shared.GetEnv("PORT", "3000"))
	gin.SetMode(gin.ReleaseMode)
	shared.LogInit()

	router := gin.New()
	router.Use(gin.Recovery())
	router.POST("/repair", repair)

	log.Error().Err(gateway.ListenAndServe(addr, router)).Msg("Failed to initialize server")
}

func repair(ctx *gin.Context) {

	clusterName := shared.GetEnv("CLUSTER_NAME", "")
	region := shared.GetEnv("REGION", "")

	kubeConfig, err := shared.GetKubeConfig(clusterName, region)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get kube config")
		ctx.PureJSON(http.StatusInternalServerError, gin.H{"message": "Failed to get kube config", "error": err.Error()})
		return
	}

	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get kube client")
		ctx.PureJSON(http.StatusInternalServerError, gin.H{"message": "Failed to get kube client", "error": err.Error()})
		return
	}

	repairRequest := &shared.RepairRequest{}
	if err := ctx.ShouldBindJSON(repairRequest); err != nil {
		log.Error().Err(err).Msg("Failed to parse json request body")
		ctx.PureJSON(http.StatusBadRequest, gin.H{"message": "Failed to parse json request body", "error": err.Error()})
		return
	}

	if repairRequest.Namespace == "" {
		repairRequest.Namespace = "default"
	}

	if repairRequest.Replicas == 0 {
		repairRequest.Replicas = 1
	}

	if repairRequest.Type == "" {
		repairRequest.Type = "deployment"
	}

	if repairRequest.Name == "" {
		log.Error().Msg("Resource name is required")
		ctx.PureJSON(http.StatusBadRequest, gin.H{"message": "Resource name is required"})
		return
	}

	var newScale *scalev1.Scale

	switch strings.ToLower(repairRequest.Type) {
	case "deployment", "deployments":
		s, err := kubeClient.AppsV1().Deployments(repairRequest.Namespace).GetScale(repairRequest.Name, metav1.GetOptions{})
		if err != nil {
			log.Error().Err(err).Msg("Failed to get resource scale")
			ctx.PureJSON(http.StatusInternalServerError, gin.H{"message": "Failed to get deployment scale", "error": err.Error()})
			return
		}
		oldScale := *s
		oldScale.Spec.Replicas = repairRequest.Replicas
		us, err := kubeClient.AppsV1().Deployments(repairRequest.Namespace).UpdateScale(repairRequest.Name, &oldScale)
		if err != nil {
			log.Error().Err(err).Msg("Failed to update resource scale")
			ctx.PureJSON(http.StatusInternalServerError, gin.H{"message": "Failed to update resource scale", "error": err.Error()})
			return
		}
		newScale = us
	case "statefulset", "statefulsets":
		s, err := kubeClient.AppsV1().StatefulSets(repairRequest.Namespace).GetScale(repairRequest.Name, metav1.GetOptions{})
		if err != nil {
			log.Error().Err(err).Msg("Failed to get resource scale")
			ctx.PureJSON(http.StatusInternalServerError, gin.H{"message": "Failed to get resource scale", "error": err.Error()})
			return
		}
		oldScale := *s
		oldScale.Spec.Replicas = repairRequest.Replicas
		us, err := kubeClient.AppsV1().StatefulSets(repairRequest.Namespace).UpdateScale(repairRequest.Name, &oldScale)
		if err != nil {
			log.Error().Err(err).Msg("Failed to update resource scale")
			ctx.PureJSON(http.StatusInternalServerError, gin.H{"message": "Failed to update resource scale", "error": err.Error()})
			return
		}
		newScale = us
	default:
		log.Error().Msg("Unsupported resource type")
		ctx.PureJSON(http.StatusBadRequest, gin.H{"message": "Unsupported resource type"})
		return
	}

	log.Info().Int32("replicas", newScale.Spec.Replicas).Msg("Resource scaled")

	ctx.PureJSON(http.StatusOK, gin.H{"message": "ok"})
}
