# partner_api

## Visão geral

O pacote `partner_api` (em `go/src/partner_api/`) simula um parceiro de pagamentos externo. Ele expõe duas operações que o restante do serviço consome:

- `Pay` — dispara uma transferência para o parceiro.
- `IsWebhookValid` — valida a assinatura HMAC dos webhooks que o parceiro envia de volta.

Você consome esse pacote — não precisa reimplementar nada dele.

## `partner_api.Pay`

```go
func Pay(ctx context.Context, req TransferRequest) (TransferResult, error)
```

Onde:

```go
type TransferRequest struct {
    ID     string
    Amount float64
}

type TransferResult struct {
    ID     string
    Status string
}
```

### Retorno em sucesso

`Pay` retorna `TransferResult{ID: <id do parceiro>, Status: partner_api.StatusConfirmed}` (`StatusConfirmed == "CONFIRMED"`) e `error == nil`. Não existe estado intermediário — ou a chamada completa com `CONFIRMED`, ou retorna erro.

### Erros possíveis

Em qualquer falha o retorno é `TransferResult{}` (zero value) e um `error` tipado ou embrulhado com `fmt.Errorf("...: %w", ...)`. Use `errors.Is` para identificar:

- `ErrPartnerUnavailable` — parceiro fora do ar ou falha de rede.
- `ErrInvalidCredentials` — secrets ausentes ou inválidos.
- `ErrBlockedDocument` — documento do destinatário bloqueado pelo parceiro.
- `ErrTransferAlreadyClosed` — transfer já foi fechada anteriormente (HTTP 409); pode ser tratado como idempotente.

Além dos erros tipados, o parceiro pode devolver códigos no corpo de erro (`ErrorBody.Code`). As constantes expostas para esses códigos são:

- `DuplicatedTransferErrorCode` — transferência duplicada.
- `BlockedDocumentErrorCode` — documento bloqueado (também mapeado para `ErrBlockedDocument`).

### Exemplo de uso

```go
result, err := partner_api.Pay(ctx, partner_api.TransferRequest{
    ID:     dep.DepositId,
    Amount: dep.Amount,
})
if err != nil {
    // tratar erro (ver lista acima)
    return err
}
// result.Status == partner_api.StatusConfirmed
```

## `partner_api.IsWebhookValid`

```go
func IsWebhookValid(signature, body string) bool
```

Valida o header `x-partner-signature` (HMAC-SHA256 do body com `apiSecret`) contra o body cru recebido. O handler HTTP em `go/src/webhook/main.go` já chama essa função antes de invocar `HandleWebhook` — você não precisa tocar nisso.

## Webhook payload

O body do webhook é desserializado em `webhook.Body` (definido em `go/src/webhook/src/domain.go`):

```go
type Body struct {
    Id      string // id do evento no parceiro
    Object  string // sempre "Deposits" para esse fluxo (constante DEPOSIT_OBJ)
    Date    string
    Data    Data
    Version string
}

type Data struct {
    Id                string  // id da transfer no parceiro
    Status            string  // ver valores abaixo
    StatusDescription *string
    IntegrationId     string  // <- depositId interno (o que você passou em TransferRequest.ID)
    BankReceiptURL    string
    AuthorizationCode *string
}
```

`Data.IntegrationId` é o `depositId` interno do nosso lado — use-o para localizar o depósito.

`Data.Status` chega como uma das três strings:

- `DONE`
- `FAILED`
- `RETURNED`

(Constantes `webhook.DONE`, `webhook.FAILED`, `webhook.RETURNED`.) Cabe a você decidir como cada um se traduz para o estado do depósito no domínio.

## O que você precisa fazer

Dois TODOs consomem esse pacote:

- `go/src/deposit_in_provider/src/pay_user.go` — chamar `partner_api.Pay` e reagir ao resultado/erros listados acima.
- `go/src/webhook/src/webhook.go` (`HandleWebhook`) — interpretar o `Body` recebido e atualizar o depósito de acordo com `Data.Status`.

Tudo o que você precisa de `partner_api` está listado neste documento; não é necessário abrir `client.go`.
