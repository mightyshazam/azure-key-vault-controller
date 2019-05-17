package azurekeyvaultsecret

import (
	"context"
	"fmt"
	secretsv1alpha1 "github.com/aware-hq/azure-key-vault-controller/pkg/apis/secrets/v1alpha1"
	kvc "github.com/aware-hq/azure-key-vault-controller/pkg/azurekeyvault/client"
	"github.com/spf13/pflag"
	"reflect"

	"github.com/Azure/azure-sdk-for-go/services/keyvault/2016-10-01/keyvault"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var (
	flagSet *pflag.FlagSet
	log      = logf.Log.WithName("controller_azurekeyvaultsecret")
	useAzureEnvironmentVariables bool
	azureConfig string
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

func init() {
	flagSet = pflag.NewFlagSet("azurekeyvault", pflag.ExitOnError)
	flagSet.BoolVar(&useAzureEnvironmentVariables, "azure-use-environment", false, "Get azure configuration from environment)")
	flagSet.StringVar(&azureConfig, "azure-config", "/etc/kubernetes/azure.json", "Location of azure config.")
}

func FlagSet() *pflag.FlagSet {
	return flagSet
}

// Add creates a new AzureKeyVaultSecret Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileAzureKeyVaultSecret{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("azurekeyvaultsecret-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource AzureKeyVaultSecret
	err = c.Watch(&source.Kind{Type: &secretsv1alpha1.AzureKeyVaultSecret{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Secrets and requeue the owner AzureKeyVaultSecret
	err = c.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &secretsv1alpha1.AzureKeyVaultSecret{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileAzureKeyVaultSecret{}

// ReconcileAzureKeyVaultSecret reconciles a AzureKeyVaultSecret object
type ReconcileAzureKeyVaultSecret struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a AzureKeyVaultSecret object and makes changes based on the state read
// and what is in the AzureKeyVaultSecret.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileAzureKeyVaultSecret) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling AzureKeyVaultSecret")

	// Fetch the AzureKeyVaultSecret instance
	instance := &secretsv1alpha1.AzureKeyVaultSecret{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Define a new Pod object
	secret, err := newSecretForCr(instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Set AzureKeyVaultSecret instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, secret, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Secret already exists
	found := &corev1.Secret{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: secret.Name, Namespace: secret.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Secret", "Secret.Namespace", secret.Namespace, "Secret.Name", secret.Name)
		err = r.client.Create(context.TODO(), secret)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Pod created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	ok := compareHashes(found, secret)
	if ok {
		// Secret already exists - don't requeue
		reqLogger.Info("Skip reconcile: Secret already exists", "Secret.Namespace", found.Namespace, "Secret.Name", found.Name)
		return reconcile.Result{}, nil
	}

	reqLogger.Info("Reconcile: Secret requires update", "Secret.Namespace", found.Namespace, "Secret.Name", found.Name)
	secret.UID = found.UID
	err = r.client.Update(context.TODO(), secret)
	if err != nil {
		return reconcile.Result{}, err
	} else {
		return reconcile.Result{}, nil
	}
}

func compareHashes(found, secret *corev1.Secret) bool {
	if !reflect.DeepEqual(found.Annotations, secret.Annotations) {
		return false
	}

	if !reflect.DeepEqual(found.Labels, secret.Labels) {
		return false
	}

	if !reflect.DeepEqual(found.Data, secret.Data) {
		return false
	}

	return true
}

func getKeysClient() (keyvault.BaseClient, error) {
	keyClient := keyvault.New()
	credentials, err := getCredentials()
	if err != nil {

		return keyClient, err
	}

	a, err := credentials.Authorizer()
	if err != nil {
		return keyClient, err
	}

	keyClient.Authorizer = a
	return keyClient, nil
}

func getCredentials() (*kvc.AzureKeyVaultCredentials, error) {
	if useAzureEnvironmentVariables {
		return kvc.NewAzureKeyVaultCredentialsFromEnvironment()
	} else {
		return kvc.NewAzureKeyVaultCredentialsFromCloudConfig(azureConfig)
	}
}

// newSecretForCr returns a secret with the same name/namespace as the cr and the content of the referenced keyvault secrets
func newSecretForCr(cr *secretsv1alpha1.AzureKeyVaultSecret) (*corev1.Secret, error) {
	kv, err := getKeysClient()
	if err != nil {
		return nil, err
	}

	var files []string
	data := map[string][]byte{}
	for _, value := range cr.Spec.Secrets {
		secret, err := kv.GetSecret(context.TODO(), cr.Spec.KeyVault, value.Key, value.Version)
		if err != nil {
			return nil, fmt.Errorf("unable to process key %s in secret %s: %v", value.Key, cr.ObjectMeta.Name, err)
		}

		if _, ok := secret.Tags["write-to-file"]; ok {
			files = append(files, value.Key)
		}

		if secret.Value != nil {
			data[value.Name] = []byte(*secret.Value)
		}
	}

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: cr.Annotations,
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    cr.ObjectMeta.Labels,
		},
		Data: data,
	}, nil
}