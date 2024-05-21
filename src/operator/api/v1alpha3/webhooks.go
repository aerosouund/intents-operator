package v1alpha3

import (
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// MySQLServerConfig //

func (in *MySQLServerConfig) SetupWebhookWithManager(mgr ctrl.Manager, validator webhook.CustomValidator) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(in).WithValidator(validator).
		Complete()
}

// PostgreSQLServerConfig //

func (in *PostgreSQLServerConfig) SetupWebhookWithManager(mgr ctrl.Manager, validator webhook.CustomValidator) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(in).WithValidator(validator).
		Complete()
}

// KafkaServerConfig //

func (ksc *KafkaServerConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(ksc).
		Complete()
}

// ProtectedService //

func (in *ProtectedService) SetupWebhookWithManager(mgr ctrl.Manager, validator webhook.CustomValidator) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(in).WithValidator(validator).
		Complete()
}

func (in *ProtectedService) Hub() {}

// ClientIntents //

func (in *ClientIntents) SetupWebhookWithManager(mgr ctrl.Manager, validator webhook.CustomValidator) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(in).WithValidator(validator).
		Complete()
}

func (in *ClientIntents) Hub() {}
