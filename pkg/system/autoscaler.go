package system

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	kedav1alpha1 "github.com/kedacore/keda/v2/apis/keda/v1alpha1"
	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/bundle"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	apiregistration "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	schema = "https://"
)

func (r *Reconciler) reconcileAutoscaler() error {
	log := r.Logger.WithField("func", "reconcileAutoscaler")
	var err error
	var autoscalerType nbv1.AutoscalerTypes = ""
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
		if r.AdapterHPA.Spec.Metrics[0].Type == autoscalingv2.ResourceMetricSourceType {
			r.Logger.Debugf("HPAV2 autoscaler type is %s, skipping HPAV2 resource creation", autoscalingv2.ResourceMetricSourceType)
			if err := r.ReconcileObject(r.AdapterHPA, r.reconcileAdapterHPA); err != nil {
				return err
			}
			return nil
		}
		prometheus, err := getPrometheus(log, prometheusNamespace)
		if err != nil {
			return err
		}
		if err = r.autoscaleHPAV2(prometheus); err != nil {
			return err
		}
	}
	return nil
}

func (r *Reconciler) cleanPreviousAutoscalingResources(log *logrus.Entry, autoscalerType nbv1.AutoscalerTypes) error {

	if err := r.ensureHPAV1Cleanup(log); err != nil {
		return err
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

func (r *Reconciler) ensureHPAV1Cleanup(log *logrus.Entry) error {
	noobaaHpaSelector, _ := labels.Parse("app=noobaa")
	noobaaHpaFieldSelector, _ := fields.ParseSelector("metadata.name=noobaa-endpoint")
	// List autoscaler only for HPAV1 based on the label and name,
	// Added name beause autoscalingv1 KubeList will fetch autoscaler with version v1 and v2
	autoscalerv1 := &autoscalingv1.HorizontalPodAutoscalerList{}
	util.KubeList(autoscalerv1, &client.ListOptions{Namespace: r.Request.Namespace, LabelSelector: noobaaHpaSelector,
		FieldSelector: noobaaHpaFieldSelector})
	for _, item := range autoscalerv1.Items {
		autoscalersName := item.Name
		log.Infof("Delete HPAV1 autoscaler Resources with name %s", autoscalersName)
		if !util.KubeDelete(&item) {
			log.Errorf("Falied to delete HPAV1 existing %q autoscaler", autoscalersName)
			return fmt.Errorf("falied to delete HPAV1 existing %q autoscaler", autoscalersName)
		}
	}
	return nil
}

func (r *Reconciler) ensureKedaCleanup(log *logrus.Entry) error {
	noobaaHpaSelector, _ := labels.Parse("app=noobaa")
	noobaaKedaFieldSelector, _ := fields.ParseSelector("metadata.name=keda-hpa-noobaa")
	// List autoscalers based on the label and name for keda ,
	// Added name beause autoscalingv2 KubeList will fetch autoscaler with version v1 and v2
	autoscalersv2 := &autoscalingv2.HorizontalPodAutoscalerList{}
	util.KubeList(autoscalersv2, &client.ListOptions{Namespace: r.Request.Namespace, LabelSelector: noobaaHpaSelector,
		FieldSelector: noobaaKedaFieldSelector})

	if len(autoscalersv2.Items) == 0 {
		return nil
	}
	log.Infof("Delete Keda autoscaler with name %s", autoscalersv2.Items[0].Name)
	deleteKedaResources(r.Request.Namespace)
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

func (r *Reconciler) ensureHPAV2Cleanup(log *logrus.Entry) error {

	noobaaHpav2Selector, _ := labels.Parse("app=noobaa")
	noobaaHPAV2FieldSelector, _ := fields.ParseSelector("metadata.name=noobaa-hpav2")
	// List autoscalers based on the label and name for HPAV2 ,
	// Added name beause autoscalingv2 KubeList will fetch autoscaler with version v1 and v2
	autoscalersv2 := &autoscalingv2.HorizontalPodAutoscalerList{}
	util.KubeList(autoscalersv2, &client.ListOptions{Namespace: r.Request.Namespace, LabelSelector: noobaaHpav2Selector,
		FieldSelector: noobaaHPAV2FieldSelector})

	for _, item := range autoscalersv2.Items {
		log.Infof("Delete HPAV2 autoscaler with name %s", item.Name)
		//  Delete HPAV2 resources, role rolebinding, serviceAccount, service, deployments, APIService etc
		if err := deleteHPAV2Resources(r.Request.Namespace); err != nil {
			log.Errorf("falied to delete HPAV2 Resources : %s", err)
			return err
		}
		if !util.KubeDelete(&item) {
			log.Errorf("Falied to delete HPAV2 existing %q autoscaler", item.Name)
			return fmt.Errorf("falied to delete HPAV2 existing %q autoscaler", item.Name)
		}
	}
	return nil
}

func deleteHPAV2Resources(namespace string) error {
	log := util.Logger().WithField("func", "deleteHPAV2Resources")
	hpav2Selector, _ := labels.Parse("app=prometheus-adapter")
	autoscalerSelector, _ := labels.Parse("origin=autoscaler")
	clusterRoleBindings := &rbacv1.ClusterRoleBindingList{}
	clusterRoles := &rbacv1.ClusterRoleList{}
	rolesBindings := &rbacv1.RoleBindingList{}
	roles := &rbacv1.RoleList{}
	hpav2Secrets := &corev1.SecretList{}
	hpav2Config := &corev1.ConfigMapList{}

	deployments := &appsv1.DeploymentList{}
	services := &corev1.ServiceList{}
	apiServices := &apiregistration.APIServiceList{}

	util.KubeList(clusterRoleBindings, &client.ListOptions{LabelSelector: hpav2Selector})
	util.KubeList(clusterRoles, &client.ListOptions{LabelSelector: hpav2Selector})
	util.KubeList(rolesBindings, &client.ListOptions{LabelSelector: hpav2Selector})
	util.KubeList(roles, &client.ListOptions{Namespace: namespace, LabelSelector: hpav2Selector})
	util.KubeList(hpav2Secrets, &client.ListOptions{Namespace: namespace, LabelSelector: autoscalerSelector})
	util.KubeList(hpav2Config, &client.ListOptions{Namespace: namespace, LabelSelector: hpav2Selector})

	util.KubeList(deployments, &client.ListOptions{Namespace: namespace, LabelSelector: hpav2Selector})
	util.KubeList(services, &client.ListOptions{Namespace: namespace, LabelSelector: hpav2Selector})
	util.KubeList(apiServices, &client.ListOptions{Namespace: namespace, LabelSelector: hpav2Selector})

	for _, obj := range clusterRoleBindings.Items {
		log.Warnf("clusterRoleBindings %q removing without grace", obj.Name)
		if !util.KubeDelete(&obj, client.GracePeriodSeconds(0)) {
			return fmt.Errorf("failed to delete clusterRoleBindings")
		}
	}
	for _, obj := range clusterRoles.Items {
		log.Warnf("clusterRoles %q removing without grace", obj.Name)
		if !util.KubeDelete(&obj, client.GracePeriodSeconds(0)) {
			return fmt.Errorf("failed to delete clusterRoles")
		}
	}
	for _, obj := range rolesBindings.Items {
		log.Warnf("rolesBindings %q removing without grace", obj.Name)
		if !util.KubeDelete(&obj, client.GracePeriodSeconds(0)) {
			return fmt.Errorf("failed to delete rolesBindings")
		}
	}
	for _, obj := range roles.Items {
		log.Warnf("roles %q removing without grace", obj.Name)
		if !util.KubeDelete(&obj, client.GracePeriodSeconds(0)) {
			return fmt.Errorf("failed to delete roles")
		}
	}

	for _, obj := range deployments.Items {
		log.Warnf("deployments %q removing without grace", obj.Name)
		if !util.KubeDelete(&obj, client.GracePeriodSeconds(0)) {
			return fmt.Errorf("failed to delete deployments")
		}
	}
	for _, obj := range services.Items {
		log.Warnf("services %q removing without grace", obj.Name)
		if !util.KubeDelete(&obj, client.GracePeriodSeconds(0)) {
			return fmt.Errorf("failed to delete services")
		}
	}
	for _, obj := range apiServices.Items {
		log.Warnf("APIServices %q removing without grace", obj.Name)
		if !util.KubeDelete(&obj, client.GracePeriodSeconds(0)) {
			return fmt.Errorf("failed to delete APIServices")
		}
	}
	for _, obj := range hpav2Secrets.Items {
		log.Warnf("HPAV2 Secrets %q removing without grace", obj.Name)
		if !util.KubeDelete(&obj, client.GracePeriodSeconds(0)) {
			return fmt.Errorf("failed to delete Secrets")
		}
	}
	for _, obj := range hpav2Config.Items {
		log.Warnf("HPAV2 ConfigMap %q removing without grace", obj.Name)
		if !util.KubeDelete(&obj, client.GracePeriodSeconds(0)) {
			return fmt.Errorf("failed to delete ConfigMap")
		}
	}
	return nil
}

func (r *Reconciler) autoscaleKeda(prometheus *monitoringv1.Prometheus) error {
	log := r.Logger.WithField("func", "autoscaleKeda")
	if !r.validateKeda() {
		log.Errorf("❌  Keda deploymen not ready")
		return errors.New("Keda deployment not ready")
	}
	log.Infof("✅  Keda found")
	if r.KedaScaled.Spec.Triggers[0].AuthenticationRef != nil {
		promethesNamespace := prometheus.Namespace
		serviceAccountName := prometheus.Spec.ServiceAccountName

		authSecretTargetRef := r.createAuthSecretTargetRef(serviceAccountName, promethesNamespace, log)
		r.KedaTriggerAuthentication.Spec.SecretTargetRef = authSecretTargetRef
		if !util.KubeCreateSkipExisting(r.KedaTriggerAuthentication) {
			log.Errorf("❌ Failed to create KedaTriggerAuthetication")
			return fmt.Errorf("Failed to create KedaTriggerAuthetication")
		}

		//create ScaledObject
		r.KedaScaled.Spec.Triggers[0].AuthenticationRef = &kedav1alpha1.ScaledObjectAuthRef{
			Name: r.KedaTriggerAuthentication.Name,
		}
		prometheusURL, err := getPrometheusURL(serviceAccountName, promethesNamespace)
		if err != nil {
			return err
		}
		r.KedaScaled.Spec.Triggers[0].Metadata["serverAddress"] = prometheusURL
		query := strings.Replace(r.KedaScaled.Spec.Triggers[0].Metadata["query"], "placeholder", r.Request.Namespace, 1)
		r.KedaScaled.Spec.Triggers[0].Metadata["query"] = query
	}
	if err := r.ReconcileObject(r.KedaScaled, r.reconcileKedaReplicaCount); err != nil {
		return err
	}
	return nil
}

func (r *Reconciler) reconcileKedaReplicaCount() error {
	endpointsSpec := r.NooBaa.Spec.Endpoints
	var minReplicas int32 = 1
	var maxReplicas int32 = 2
	if endpointsSpec != nil {
		minReplicas = endpointsSpec.MinCount
		maxReplicas = endpointsSpec.MaxCount
	}
	r.KedaScaled.Spec.MinReplicaCount = &minReplicas
	r.KedaScaled.Spec.MaxReplicaCount = &maxReplicas
	return nil
}

func (r *Reconciler) createAuthSecretTargetRef(serviceAccountName, promethesNamespace string, log *logrus.Entry) []kedav1alpha1.AuthSecretTargetRef {
	var authSecretTargetRef []kedav1alpha1.AuthSecretTargetRef
	err := r.checkAndCreatePrometheusSecret(serviceAccountName, promethesNamespace)
	if err != nil {
		// changes to work with minikube deployment
		log.Warnf("❌ Failed to get Prometheus Secret: %s", err)
		authSecretTargetRef = r.creatBasicSecretTargetRef(log)
	} else {
		r.KedaScaled.Spec.Triggers[0].Metadata["authModes"] = "bearer"
		authSecretTargetRef = []kedav1alpha1.AuthSecretTargetRef{
			{
				Parameter: "bearerToken",
				Name:      "prometheus-k8s-secret",
				Key:       "token",
			},
			{
				Parameter: "ca",
				Name:      "prometheus-k8s-secret",
				Key:       "service-ca.crt",
			},
		}
	}
	return authSecretTargetRef
}

func (r *Reconciler) creatBasicSecretTargetRef(log *logrus.Entry) []kedav1alpha1.AuthSecretTargetRef {
	kedaSecretObject := util.KubeObject(bundle.File_deploy_internal_hpa_keda_secret_yaml).(*corev1.Secret)
	kedaSecretObject.Namespace = r.Request.Namespace
	kedaSecretObject.Data = map[string][]byte{
		"username": []byte("username"),
	}
	if !util.KubeCreateSkipExisting(kedaSecretObject) {
		log.Errorf("❌ dummy prometheus secret not created %s", kedaSecretObject)
	}

	r.KedaScaled.Spec.Triggers[0].Metadata["unsafeSsl"] = "true"
	r.KedaScaled.Spec.Triggers[0].Metadata["authModes"] = "basic"
	authSecretTargetRef := []kedav1alpha1.AuthSecretTargetRef{
		{
			Parameter: "username",
			Name:      "prometheus-k8s-secret",
			Key:       "username",
		},
	}
	return authSecretTargetRef
}

func (r *Reconciler) validateKeda() bool {
	err := r.Client.Get(r.Ctx, util.ObjectKey(r.KedaScaled), r.KedaScaled)
	return !(meta.IsNoMatchError(err) || runtime.IsNotRegisteredError(err))
}

func getPrometheus(log *logrus.Entry, prometheusNamespace string) (*monitoringv1.Prometheus, error) {
	prometheusList := &monitoringv1.PrometheusList{}
	if !util.KubeList(prometheusList, &client.ListOptions{Namespace: prometheusNamespace}) {
		return nil, fmt.Errorf("prometheus not found in namespace %s", prometheusNamespace)
	}
	if len(prometheusList.Items) == 0 {
		log.Errorf("❌ prometheus not found in namespace %s", prometheusNamespace)
		return nil, fmt.Errorf("prometheus not found in namespace %s", prometheusNamespace)
	}
	return prometheusList.Items[0], nil
}

func (r *Reconciler) checkAndCreatePrometheusSecret(serviceAccountName string, promethesNamespace string) error {
	secrets := &corev1.SecretList{}
	util.KubeList(secrets, &client.ListOptions{Namespace: promethesNamespace})
	if len(secrets.Items) == 0 {
		return fmt.Errorf("prometheus secret not found")
	}
	kedaSecretObject := &corev1.Secret{}
	for _, secret := range secrets.Items {
		if strings.Contains(secret.Name, serviceAccountName+"-token") {
			kedaSecretObject = util.KubeObject(bundle.File_deploy_internal_hpa_keda_secret_yaml).(*corev1.Secret)
			kedaSecretObject.Namespace = r.Request.Namespace
			kedaSecretObject.Data = secret.Data
		}
	}
	if kedaSecretObject.Namespace == "" {
		return fmt.Errorf("prometheus secret not created")
	}
	util.KubeCreateSkipExisting(kedaSecretObject)
	return nil
}

func deleteKedaResources(namespace string) {
	log := util.Logger()
	kedaSelector, _ := labels.Parse("origin=keda")
	scaledObjects := &kedav1alpha1.ScaledObjectList{}
	triggerAuthentications := &kedav1alpha1.TriggerAuthenticationList{}
	kedaSecrets := &corev1.SecretList{}
	util.KubeList(scaledObjects, &client.ListOptions{Namespace: namespace, LabelSelector: kedaSelector})
	util.KubeList(triggerAuthentications, &client.ListOptions{Namespace: namespace, LabelSelector: kedaSelector})
	util.KubeList(kedaSecrets, &client.ListOptions{Namespace: namespace, LabelSelector: kedaSelector})

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

func getPrometheusURL(serviceAccountName, promethesNamespace string) (string, error) {
	prometheusService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceAccountName,
			Namespace: promethesNamespace,
		},
	}
	klient := util.KubeClient()
	var prometheusPort string
	if err := klient.Get(util.Context(), util.ObjectKey(prometheusService), prometheusService); err != nil {
		log.Printf("❌ Failed to get Prometheus service: %s", err)
		return "", err
	}
	for _, port := range prometheusService.Spec.Ports {
		if strings.Contains(port.Name, "web") {
			prometheusPort = strconv.Itoa(int(port.Port))
		}
	}
	if prometheusPort == "" {
		log.Println("❌ Failed to find the promethues port")
		return "", fmt.Errorf("Failed to find the promethues port")
	}
	prometheusURL := schema + serviceAccountName + "." + promethesNamespace + ".svc:" + prometheusPort
	return prometheusURL, nil
}

func (r *Reconciler) autoscaleHPAV2(prometheus *monitoringv1.Prometheus) error {
	log := r.Logger.WithField("func", "autoscaleHPAV2")
	log.Infof("✅  Start autoscaling with HPAV2")
	if err := r.reconcileHPAV2RBAC(); err != nil {
		log.Println("❌ Failed create HPAV2 RBAC", err)
		return err
	}
	if err := r.reconcilePrometheusAdapterResources(prometheus); err != nil {
		log.Println("❌ Failed create HPAV2 Adapter", err)
		return err
	}
	return nil
}

func (r *Reconciler) reconcilePrometheusAdapterResources(prometheus *monitoringv1.Prometheus) error {

	promethesNamespace := prometheus.Namespace
	serviceAccountName := prometheus.Spec.ServiceAccountName
	prometheusURL, err := getPrometheusURL(serviceAccountName, promethesNamespace)
	if err != nil {
		return err
	}
	adapterServingCertsConfigMap := util.KubeObject(bundle.File_deploy_internal_hpav2_serving_certs_ca_bundle_yaml).(*corev1.ConfigMap)
	adapterServingCertsConfigMap.Namespace = r.Request.Namespace
	if err := r.ReconcileObject(adapterServingCertsConfigMap, nil); err != nil {
		return err
	}

	adapterAuthConfigMap := util.KubeObject(bundle.File_deploy_internal_hpav2_prometheus_auth_config_yaml).(*corev1.ConfigMap)
	adapterAuthConfigMap.Namespace = r.Request.Namespace
	prometheusConfig := adapterAuthConfigMap.Data["prometheus-config.yaml"]
	adapterAuthConfigMap.Data["prometheus-config.yaml"] = strings.Replace(prometheusConfig, "prometheus-url-placeholder", prometheusURL, 1)
	if err := r.ReconcileObject(adapterAuthConfigMap, nil); err != nil {
		return err
	}

	adapterConfigMap := util.KubeObject(bundle.File_deploy_internal_hpav2_configmap_adapter_yaml).(*corev1.ConfigMap)
	adapterConfigMap.Namespace = r.Request.Namespace
	config := adapterConfigMap.Data["config.yaml"]
	adapterConfigMap.Data["config.yaml"] = strings.Replace(config, "placeholder", r.Request.Namespace, 1)
	if err := r.ReconcileObject(adapterConfigMap, nil); err != nil {
		return err
	}
	adapterDeployment := util.KubeObject(bundle.File_deploy_internal_hpav2_deployment_adapter_yaml).(*appsv1.Deployment)
	adapterDeployment.Namespace = r.Request.Namespace
	if waitForCertificateReady(adapterServingCertsConfigMap) && adapterServingCertsConfigMap.Data["service-ca.crt"] != "" {
		adapterDeployment.Spec.Template.Spec.Containers[0].Args = append(adapterDeployment.Spec.Template.Spec.Containers[0].Args, "--prometheus-url="+prometheusURL,
			"--prometheus-auth-config=/etc/prometheus-config/prometheus-config.yaml", "--tls-cert-file=/var/run/serving-cert/tls.crt",
			"--tls-private-key-file=/var/run/serving-cert/tls.key")

	} else {
		adapterDeployment.Spec.Template.Spec.Containers[0].Args = append(adapterDeployment.Spec.Template.Spec.Containers[0].Args, "--prometheus-url="+prometheusURL,
			"--cert-dir=/var/run/empty/serving-cert")
	}

	if err := r.ReconcileObject(adapterDeployment, nil); err != nil {
		return err
	}
	adapterService := util.KubeObject(bundle.File_deploy_internal_hpav2_service_adapter_yaml).(*corev1.Service)
	adapterService.Namespace = r.Request.Namespace
	if err := r.ReconcileObject(adapterService, nil); err != nil {
		return err
	}
	adapterAPIService := util.KubeObject(bundle.File_deploy_internal_hpav2_apiservice_yaml).(*apiregistration.APIService)
	adapterAPIService.Spec.Service.Namespace = r.Request.Namespace
	if !util.KubeCreateSkipExisting(adapterAPIService) {
		return fmt.Errorf("Error while crateiing APIService prometheus-adapter")
	}
	if err := r.ReconcileObject(r.AdapterHPA, r.reconcileAdapterHPA); err != nil {
		return err
	}
	return nil
}

func (r *Reconciler) reconcileAdapterHPA() error {
	// this method will update the noobaa endpoint HPA min and max replicas count with respect to
	// noobaa CR endpoint min and max values
	endpointsSpec := r.NooBaa.Spec.Endpoints
	var minReplicas int32 = 1
	var maxReplicas int32 = 2
	if endpointsSpec != nil {
		minReplicas = endpointsSpec.MinCount
		maxReplicas = endpointsSpec.MaxCount
	}
	r.AdapterHPA.Spec.MinReplicas = &minReplicas
	r.AdapterHPA.Spec.MaxReplicas = maxReplicas
	// target value should be nil otherwise Kubernetes HPA reconciler tries to validate the value,
	// for type AverageValue this validation is not required.
	if r.AdapterHPA.Spec.Metrics[0].Object != nil {
		r.AdapterHPA.Spec.Metrics[0].Object.Target.Value = nil
	}
	return nil
}

func (r *Reconciler) reconcileHPAV2RBAC() error {
	serviceAccount := util.KubeObject(bundle.File_deploy_service_acount_hpav2_yaml).(*corev1.ServiceAccount)
	serviceAccount.Namespace = r.Request.Namespace
	if err := r.ReconcileObject(serviceAccount, nil); err != nil {
		return err
	}
	roleResourceReader := util.KubeObject(bundle.File_deploy_role_resource_reader_hpav2_yaml).(*rbacv1.ClusterRole)
	if !util.KubeCreateSkipExisting(roleResourceReader) {
		return fmt.Errorf("Error while crateiing ClusterRole prometheus-adapter-resource-reader")
	}
	roleServerResources := util.KubeObject(bundle.File_deploy_role_server_resources_hpav2_yaml).(*rbacv1.Role)
	roleServerResources.Namespace = r.Request.Namespace
	if err := r.ReconcileObject(roleServerResources, nil); err != nil {
		return err
	}
	roleBindingAuthDelegator := util.KubeObject(bundle.File_deploy_role_binding_auth_delegator_hpav2_yaml).(*rbacv1.ClusterRoleBinding)
	roleBindingAuthDelegator.Subjects[0].Namespace = r.Request.Namespace
	if !util.KubeCreateSkipExisting(roleBindingAuthDelegator) {
		return fmt.Errorf("Error while crateiing ClusterRoleBinding prometheus-adapter-system-auth-delegator")
	}
	roleBindingAuthReader := util.KubeObject(bundle.File_deploy_role_binding_auth_reader_hpav2_yaml).(*rbacv1.RoleBinding)
	roleBindingAuthReader.Subjects[0].Namespace = r.Request.Namespace
	if !util.KubeCreateSkipExisting(roleBindingAuthReader) {
		return fmt.Errorf("Error while crateiing RoleBinding prometheus-adapter-auth-reader")
	}
	roleBindingResourceReader := util.KubeObject(bundle.File_deploy_role_binding_resource_reader_hpav2_yaml).(*rbacv1.ClusterRoleBinding)
	roleBindingResourceReader.Subjects[0].Namespace = r.Request.Namespace
	if !util.KubeCreateSkipExisting(roleBindingResourceReader) {
		return fmt.Errorf("Error while crateiing ClusterRoleBinding prometheus-adapter-resource-reader")
	}
	roleBindingServerResources := util.KubeObject(bundle.File_deploy_role_binding_server_resources_hpav2_yaml).(*rbacv1.RoleBinding)
	roleBindingServerResources.Namespace = r.Request.Namespace
	roleBindingServerResources.Subjects[0].Namespace = r.Request.Namespace
	if err := r.ReconcileObject(roleBindingServerResources, nil); err != nil {
		return err
	}

	return nil
}

// WaitReady waits until the system phase changes to ready by the operator
func waitForCertificateReady(configMap *corev1.ConfigMap) bool {
	log := util.Logger()
	klient := util.KubeClient()

	interval := time.Duration(3)
	timeout := time.Duration(15)

	err := wait.PollUntilContextTimeout(ctx, interval*time.Second, timeout*time.Second, true, func(ctx context.Context) (bool, error) {
		if err := klient.Get(util.Context(), util.ObjectKey(configMap), configMap); err != nil {
			log.Printf("⏳ Failed to get service ca certificate: %s", err)
			return false, nil
		}
		if configMap.Data["service-ca.crt"] == "" {
			return false, fmt.Errorf("service ca certificate not created")
		}

		return true, nil
	})
	return (err == nil)
}
