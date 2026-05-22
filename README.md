# Instant Deposits — Exercício de entrevista

Serviço Go serverless usado como exercício técnico em entrevistas. O candidato completa o ciclo de vida de um depósito (pagamento no parceiro + webhook de retorno) sobre um esqueleto pronto.

- **Brief do candidato:** [DESAFIO.md](./DESAFIO.md)
- **Código:** [`go/`](./go) — CDK + lambdas Go + DynamoDB Local

## Como rodar

```bash
cd go
npm install
npm test            # Go + DynamoDB Local
npm run test:cdk    # CDK
```

Ou abra no GitHub Codespaces — o devcontainer já configura Go, Node, Java, AWS CLI e DynamoDB Local.
