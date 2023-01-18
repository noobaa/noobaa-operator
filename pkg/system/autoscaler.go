package system

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	kedav1alpha1 "github.com/kedacore/keda/v2/apis/keda/v1alpha1"
	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/bundle"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/sirupsen/logrus"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	schema = "https://"
)

func (r *Reconciler) reconcileAutoscaler() error {
	log := r.Logger.WithField("func", "reconcileAutoscaler")
	var err error
	autoscalerType := nbv1.AutoscalerTypes(options.DefaultAutoscalerType)
	if r.NooBaa.Spec.Autoscaler.AutoscalerType != "" {
		autoscalerType = r.NooBaa.Spec.Autoscaler.AutoscalerType
	}
	log.Infof("Configured autoscaler types is %s", autoscalerType)
	if err = r.cleanPreviousAutoscalingResources(log, autoscalerType); err != nil {
		return err
	}
	prometheusNamespace := options.PrometheusNamespace
	if r.NooBaa.Spec.Autoscaler.PrometheusNamespace != "" {
		prometheusNamespace = r.NooBaa.Spec.Autoscaler.PrometheusNamespace
	}
	switch autoscalerType {
	case nbv1.AutoscalerTypeKeda:
		prometheus, err := getPrometheus(log, prometheusNamespace)
		if err != nil {
			return err
		}
		if err = r.autoscaleKeda(prometheus); err != nil {
			return err
		}
	case nbv1.AutoscalerTypeHPAV2:
		//TODO
		log.Infof("HPAV2 implementation pending")
	case nbv1.AutoscalerTypeHPAV1:
		if err = r.ReconcileObject(r.HPAEndpoint, r.SetDesiredHPAEndpoint); err != nil {
			return err
		}
	default:
		log.Errorf("❌ Autoscaler type is not defined for noobaa")
		return errors.New("Autoscaler type is not defined")
	}
	return nil
}

func (r *Reconciler) cleanPreviousAutoscalingResources(log *logrus.Entry, autoscalerType nbv1.AutoscalerTypes) error {

	if autoscalerType != nbv1.AutoscalerTypeHPAV1 {
		if err := r.ensureHPAV1Cleanup(log); err != nil {
			return err
		}
	}
	if autoscalerType != nbv1.AutoscalerTypeHPAV2 {
		if err := r.ensureHPAV2Cleanup(log); err != nil {
			return err
		}
	}
	if autoscalerType != nbv1.AutoscalerTypeKeda {
		if err := r.ensureKedaCleanup(log); err != nil {
			return err
		}
	}
	return nil
}

func (r *Reconciler) ensureKedaCleanup(log *logrus.Entry) error {
	noobaaHpaSelector, _ := labels.Parse("app=noobaa")
	noobaaKedaFieldSelector, _ := fields.ParseSelector("metadata.name=keda-hpa-noobaa-endpoint")
	// List autoscalers based on the label and name for keda ,
	// Added name beause autoscalingv2 KubeList will fetch autoscaler with version v1 and v2
	autoscalersv2 := &autoscalingv2.HorizontalPodAutoscalerList{}
	util.KubeList(autoscalersv2, &client.ListOptions{Namespace: r.Request.Namespace, LabelSelector: noobaaHpaSelector,
		FieldSelector: noobaaKedaFieldSelector})

	if len(autoscalersv2.Items) == 0 {
		return nil
	}
	log.Infof("Delete Keda autoscaler with name %s", autoscalersv2.Items[0].Name)
	deleteKedaResources()
	if !util.KubeDelete(&autoscalersv2.Items[0]) {
		log.Errorf("Falied to delete Keda existing %q autoscaler", autoscalersv2.Items[0].Name)
		return fmt.Errorf("Falied to delete Keda existing %q autoscaler", autoscalersv2.Items[0].Name)
	}
	autoscalersv2 = &autoscalingv2.HorizontalPodAutoscalerList{}
	util.KubeList(autoscalersv2, &client.ListOptions{Namespace: r.Request.Namespace, LabelSelector: noobaaHpaSelector,
		FieldSelector: noobaaKedaFieldSelector})
	if len(autoscalersv2.Items) > 0 {
		// Return error if Keda autoscaler not deleted
		return fmt.Errorf("Keda autoscaler still not deleted")
	}

	return nil
}

func (r *Reconciler) ensureHPAV1Cleanup(log *logrus.Entry) error {
	noobaaHpaSelector, _ := labels.Parse("app=noobaa")
	noobaaHpaFieldSelector, _ := fields.ParseSelector("metadata.name=noobaa-endpoint")
	// List autoscaler only for HPAV1 based on the label and name,
	// Added name beause autoscalingv1 KubeList will fetch autoscaler with version v1 and v2
	autoscalersv1 := &autoscalingv1.HorizontalPodAutoscalerList{}
	util.KubeList(autoscalersv1, &client.ListOptions{Namespace: r.Request.Namespace, LabelSelector: noobaaHpaSelector,
		FieldSelector: noobaaHpaFieldSelector})
	if len(autoscalersv1.Items) == 0 {
		return nil
	}
	autoscalersName := autoscalersv1.Items[0].Name
	log.Infof("Delete HPAV1 autoscaler Resources with name %s", autoscalersName)
	if !util.KubeDelete(&autoscalersv1.Items[0]) {
		log.Errorf("Falied to delete HPAV1 existing %q autoscaler", autoscalersName)
		return fmt.Errorf("Falied to delete HPAV1 existing %q autoscaler", autoscalersName)
	}
	autoscalersv1 = &autoscalingv1.HorizontalPodAutoscalerList{}
	util.KubeList(autoscalersv1, &client.ListOptions{Namespace: r.Request.Namespace, LabelSelector: noobaaHpaSelector,
		FieldSelector: noobaaHpaFieldSelector})
	if len(autoscalersv1.Items) > 0 {
		// Return error if HPAV1 autoscaler not deleted
		return fmt.Errorf("HPAV1 autoscaler still not deleted ")
	}
	return nil
}

func (r *Reconciler) ensureHPAV2Cleanup(log *logrus.Entry) error {
	log.Infof("HPAV2 implementation pending")

	return nil
}

func (r *Reconciler) autoscaleKeda(prometheus *monitoringv1.Prometheus) error {
	log := r.Logger.WithField("func", "autoscaleKeda")
	if !r.validateKeda() {
		log.Errorf("❌  Keda deploymen not ready")
		return errors.New("Keda deploymen not ready")
	}
	log.Infof("✅  Keda found")
	promethesNamespace := prometheus.Namespace
	serviceAccountName := prometheus.Spec.ServiceAccountName
	err := r.checkAndCreatePrometheusSecret(serviceAccountName, promethesNamespace)
	if err != nil {
		return err
	}
	authSecretTargetRef := []kedav1alpha1.AuthSecretTargetRef{
		{
			Parameter: "bearerToken",
			Name:      "prometheus-k8s-token",
			Key:       "token",
		},
		{
			Parameter: "ca",
			Name:      "prometheus-k8s-token",
			Key:       "service-ca.crt",
		},
	}
	//create TriggerAuthentication
	r.KedaTriggerAuthentication.Spec.SecretTargetRef = authSecretTargetRef
	if !util.KubeCreateSkipExisting(r.KedaTriggerAuthentication) {
		log.Errorf("❌ Failed to create KedaTriggerAuthetication")
		return fmt.Errorf("Failed to create KedaTriggerAuthetication")
	}

	//create ScaledObject
	r.KedaScaled.Spec.Triggers[0].AuthenticationRef = &kedav1alpha1.ScaledObjectAuthRef{
		Name: r.KedaTriggerAuthentication.Name,
	}

	prometheusService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceAccountName,
			Namespace: promethesNamespace,
		},
	}

	klient := util.KubeClient()
	var prometheusPort string
	if err = klient.Get(util.Context(), util.ObjectKey(prometheusService), prometheusService); err != nil {
		log.Errorf("❌ Failed to get Prometheus service: %s", err)
		return err
	}
	// fetch prometheus port for consructing the Prometheus URL form prometheus service
	for _, port := range prometheusService.Spec.Ports {
		if strings.Contains(port.Name, "web") {
			prometheusPort = strconv.Itoa(int(port.Port))
		}
	}
	if prometheusPort == "" {
		log.Errorf("❌ Failed to find the promethues port")
		return fmt.Errorf("Failed to find the promethues port")
	}
	r.KedaScaled.Spec.Triggers[0].Metadata["serverAddress"] = schema + serviceAccountName + "." + promethesNamespace + ".svc.cluster.local:" + prometheusPort
	query := strings.Replace(r.KedaScaled.Spec.Triggers[0].Metadata["query"], "placeholder", r.Request.Namespace, 1)
	r.KedaScaled.Spec.Triggers[0].Metadata["query"] = query
	endpointsSpec := r.NooBaa.Spec.Endpoints
	if endpointsSpec != nil {
		r.KedaScaled.Spec.MinReplicaCount = &endpointsSpec.MinCount
		r.KedaScaled.Spec.MaxReplicaCount = &endpointsSpec.MaxCount
	}
	if !util.KubeCreateSkipExisting(r.KedaScaled) {
		log.Errorf("❌ Failed to create KedaScaledObject")
		return fmt.Errorf("Failed to create KedaScaledObject")
	}
	return nil
}

func (r *Reconciler) validateKeda() bool {
	err := r.Client.Get(r.Ctx, util.ObjectKey(r.KedaScaled), r.KedaScaled)
	return !(meta.IsNoMatchError(err) || runtime.IsNotRegisteredError(err))
}

func getPrometheus(log *logrus.Entry, prometheusNamespace string) (*monitoringv1.Prometheus, error) {
	prometheusList := &monitoringv1.PrometheusList{}
	if !util.KubeList(prometheusList, &client.ListOptions{Namespace: prometheusNamespace}) {
		return nil, fmt.Errorf("No prometheus found in namespace %s", prometheusNamespace)
	}
	if len(prometheusList.Items) == 0 {
		log.Errorf("❌  No prometheus found in namespace %s", prometheusNamespace)
		return nil, fmt.Errorf("No prometheus found in namespace %s", prometheusNamespace)
	}
	return prometheusList.Items[0], nil
}

func (r *Reconciler) checkAndCreatePrometheusSecret(serviceAccountName string, promethesNamespace string) error {
	secrets := &corev1.SecretList{}
	util.KubeList(secrets, &client.ListOptions{Namespace: promethesNamespace})
	if len(secrets.Items) == 0 {
		return fmt.Errorf("No prometheus secret found")
	}
	for _, secret := range secrets.Items {
		if strings.Contains(secret.Name, serviceAccountName+"-token") {
			kedaSecret := util.KubeObject(bundle.File_deploy_internal_hpa_keda_secret_yaml)
			kedaSecretObject := kedaSecret.(*corev1.Secret)
			kedaSecretObject.Namespace = r.Request.Namespace
			kedaSecretObject.Data = secret.Data
			util.KubeCreateSkipExisting(kedaSecretObject)
		}
	}
	return nil
}

func deleteKedaResources() {
	log := util.Logger()
	kedaSelector, _ := labels.Parse("origin=keda")
	scaledObjects := &kedav1alpha1.ScaledObjectList{}
	triggerAuthentications := &kedav1alpha1.TriggerAuthenticationList{}
	kedaSecrets := &corev1.SecretList{}
	util.KubeList(scaledObjects, &client.ListOptions{LabelSelector: kedaSelector})
	util.KubeList(triggerAuthentications, &client.ListOptions{LabelSelector: kedaSelector})
	util.KubeList(kedaSecrets, &client.ListOptions{LabelSelector: kedaSelector})

	for _, obj := range scaledObjects.Items {
		log.Warnf("scaledObject %q removing without grace", obj.Name)
		util.KubeDelete(&obj, client.GracePeriodSeconds(0))
	}

	for _, obj := range triggerAuthentications.Items {
		log.Warnf("triggerAuthentications %q removing without grace", obj.Name)
		util.KubeDelete(&obj, client.GracePeriodSeconds(0))
	}
	for _, obj := range kedaSecrets.Items {
		log.Warnf("keda Secrets %q removing without grace", obj.Name)
		util.KubeDelete(&obj, client.GracePeriodSeconds(0))
	}
}
