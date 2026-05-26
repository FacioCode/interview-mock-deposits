# Instant Deposits — Exercício de entrevista

Serviço Go serverless usado como exercício técnico em entrevistas. O candidato completa o ciclo de vida de um depósito (pagamento no parceiro + webhook de retorno) sobre um esqueleto pronto.

- **Brief do candidato:** [DESAFIO.md](./DESAFIO.md)
- **Código:** [`go/`](./go) — Go + AWS SDK v1, CDK TypeScript, DynamoDB Local

## Como rodar

```bash
cd go
./test.sh                                   # boot DynamoDB Local + roda os testes do app
cd cdk && npm install && npm run test:cdk   # testes da stack CDK (TS)
```

Ou abra no GitHub Codespaces — o devcontainer já configura Go, Node, Java, AWS CLI e DynamoDB Local.

> Codespaces exige estar logado no GitHub antes de criar o ambiente.
