# signdocsbrasil-go

SDK oficial em Go para a API SignDocsBrasil.

## Requisitos

- Go 1.21+
- Zero dependências externas (usa apenas a biblioteca padrão)

## Instalação

```bash
go get github.com/signdocsbrasil/signdocsbrasil-go
```

## Início Rápido

```go
package main

import (
    "context"
    "fmt"
    "log"

    signdocs "github.com/signdocsbrasil/signdocsbrasil-go"
)

func main() {
    client, err := signdocs.NewClient("seu_client_id",
        signdocs.WithClientSecret("seu_client_secret"),
    )
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    tx, err := client.Transactions.Create(ctx, &signdocs.CreateTransactionRequest{
        Purpose: signdocs.TransactionPurposeDocumentSignature,
        Policy:  signdocs.Policy{Profile: signdocs.PolicyProfileClickOnly},
        Signer: signdocs.Signer{
            Name:           "João Silva",
            Email:          "joao@example.com",
            UserExternalID: "user-001",
        },
        Document: &signdocs.DocumentInline{
            Content:  pdfBase64,
            Filename: "contrato.pdf",
        },
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(tx.TransactionID, tx.Status)
}
```

### Private Key JWT (ES256)

```go
keyPem, _ := os.ReadFile("./private-key.pem")
key, _ := signdocs.ParseES256PrivateKeyFromPEM(keyPem)

client, _ := signdocs.NewClient("seu_client_id",
    signdocs.WithPrivateKey(key, "seu-key-id"),
)
```

## Recursos Disponíveis

| Recurso | Métodos |
|---------|---------|
| `client.Transactions` | `Create`, `List`, `Get`, `Cancel`, `Finalize`, `ListAutoPaginate` |
| `client.Documents` | `Upload`, `Presign`, `Confirm`, `Download` |
| `client.Steps` | `List`, `Start`, `Complete` |
| `client.Signing` | `Prepare`, `Complete` |
| `client.Evidence` | `Get` |
| `client.Verification` | `Verify`, `Downloads` |
| `client.Users` | `Enroll` |
| `client.Webhooks` | `Register`, `List`, `Delete`, `Test` |
| `client.SigningSessions` | `Create`, `GetStatus`, `Cancel`, `List`, `WaitForCompletion` |
| `client.Envelopes` | `Create`, `Get`, `AddSession`, `CombinedStamp` |
| `client.DocumentGroups` | `CombinedStamp` |
| `client.Health` | `Check`, `History` |

## Envelopes (Múltiplos Signatários)

### Criando um Envelope

```go
env, err := client.Envelopes.Create(ctx, &signdocs.CreateEnvelopeRequest{
    SigningMode:  "PARALLEL",
    TotalSigners: 2,
    Document: signdocs.EnvelopeDocument{
        Content:  pdfBase64,
        Filename: "contrato.pdf",
    },
})
if err != nil {
    log.Fatal(err)
}

fmt.Println(env.EnvelopeID, env.Status)
```

### Adicionando Sessões de Assinatura

```go
session1, err := client.Envelopes.AddSession(ctx, env.EnvelopeID, &signdocs.AddEnvelopeSessionRequest{
    Signer: signdocs.EnvelopeSessionSigner{
        Name:           "João Silva",
        Email:          "joao@example.com",
        UserExternalID: "user-001",
    },
    Policy:      signdocs.EnvelopeSessionPolicy{Profile: "CLICK_ONLY"},
    Purpose:     "DOCUMENT_SIGNATURE",
    SignerIndex: 1,
})
if err != nil {
    log.Fatal(err)
}

session2, err := client.Envelopes.AddSession(ctx, env.EnvelopeID, &signdocs.AddEnvelopeSessionRequest{
    Signer: signdocs.EnvelopeSessionSigner{
        Name:           "Maria Santos",
        Email:          "maria@example.com",
        UserExternalID: "user-002",
    },
    Policy:      signdocs.EnvelopeSessionPolicy{Profile: "CLICK_ONLY"},
    Purpose:     "DOCUMENT_SIGNATURE",
    SignerIndex: 2,
})
if err != nil {
    log.Fatal(err)
}

fmt.Println(session1.URL, session2.URL)
```

## Configuração Avançada

### HTTP Client customizado

```go
httpClient := &http.Client{
    Timeout: 20 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:    10,
        IdleConnTimeout: 30 * time.Second,
    },
}

client, _ := signdocs.NewClient("seu_client_id",
    signdocs.WithClientSecret("seu_client_secret"),
    signdocs.WithHTTPClient(httpClient),
)
```

### Logging

O SDK suporta logging estruturado via `log/slog` (Go 1.21+). São logados apenas: método HTTP, path, status code e duração. Headers de autorização, corpos de request/response e tokens nunca são logados.

```go
import "log/slog"

logger := slog.Default()

client, _ := signdocs.NewClient("seu_client_id",
    signdocs.WithClientSecret("seu_client_secret"),
    signdocs.WithLogger(logger),
)
```

### Timeout por requisição

Todas as operações aceitam `context.Context` — use `context.WithTimeout` para controlar o timeout individualmente:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

tx, err := client.Transactions.Get(ctx, "tx_123")
```

## Documentação

Para guias completos de integração com exemplos passo-a-passo de todos os fluxos de assinatura, veja a [documentação centralizada](../docs/README.md).
