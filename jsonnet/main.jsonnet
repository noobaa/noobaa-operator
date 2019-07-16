local noobaa = (import 'noobaa-mixins/mixin.libsonnet');
local prometheusRule(storageNamespace, storageName, alertType, prometheusLabel) = {
  apiVersion: 'monitoring.coreos.com/v1',
  kind: 'PrometheusRule',
  metadata: {
    labels: {
      prometheus: prometheusLabel,
      role: 'alert-rules',
    },
    name: 'prometheus-' + storageName + '-rules',
    namespace: storageNamespace,
  },
  spec: alertType,
};

{
  'noobaa-prometheus-rules': prometheusRule('default', 'noobaa', noobaa.prometheusAlerts+noobaa.prometheusRules, 'k8s'),
}
