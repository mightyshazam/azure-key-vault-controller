package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AzureKeyVaultSecretSpec defines the desired state of AzureKeyVaultSecret
// +k8s:openapi-gen=true
type AzureKeyVaultSecretSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	KeyVault string                     `json:"keyvault"`
	Secrets  []AzureKeyVaultSecretEntry `json:"secrets"`
}

type AzureKeyVaultSecretEntry struct {
	Key string `json:"key"`
	// +optional
	Name string `json:"name,omitempty"`
	// +optional
	Version string `json:"version,omitempty"`
	// +optional
	WriteToFile bool `json:"isfile,omitempty"`
}

// AzureKeyVaultSecretStatus defines the observed state of AzureKeyVaultSecret
// +k8s:openapi-gen=true
type AzureKeyVaultSecretStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AzureKeyVaultSecret is the Schema for the azurekeyvaultsecrets API
// +k8s:openapi-gen=true
type AzureKeyVaultSecret struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AzureKeyVaultSecretSpec   `json:"spec,omitempty"`
	Status AzureKeyVaultSecretStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AzureKeyVaultSecretList contains a list of AzureKeyVaultSecret
type AzureKeyVaultSecretList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AzureKeyVaultSecret `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AzureKeyVaultSecret{}, &AzureKeyVaultSecretList{})
}
