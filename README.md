# azure-key-vault-controller
A controller to copy keyvault entries to kubernetes secrets

The custom resource `AzureKeyVaultSecret` represents a secret in kubernetes that refrerences and Azure Key Vault. There is a one to one mapping between an entry in the resource and a secret in an Azure Key Vault.

## Entry Specification
| Key | Type | Required | Description |
| --- | ---- | -------- | ----------- |
| isfile | boolean | no | Indicates whether a secret should be written to a file. Captured in the annotation `secrets.awarehq.com/write-to-file` on the kubernetes secret. |
| key | string | yes | Name of corresponding Azure Key Vault entry. |
| name | string | no | Name of kubernetes secret entry. |
| version | string | no | Azure Key Vault secret version. Defaults to latest. |
