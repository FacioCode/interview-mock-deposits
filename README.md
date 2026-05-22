# Instant Deposits — Exercício de entrevista

Serviço serverless usado como exercício técnico em entrevistas. O candidato completa o ciclo de vida de um depósito (pagamento no parceiro + webhook de retorno) sobre um esqueleto pronto. Disponível em mais de uma linguagem — escolha a do seu stack.

- **Brief do candidato:** [DESAFIO.md](./DESAFIO.md)
- **Linguagens disponíveis:**
  - [`go/`](./go) — Go + AWS SDK v1, CDK TypeScript, DynamoDB Local
  - [`python/`](./python) — Python 3.12 + boto3 + uv + pytest, CDK TypeScript, DynamoDB Local

## Como rodar

```bash
cd <lang>         # go ou python
./test.sh         # boot DynamoDB Local + roda os testes do app

cd cdk && npm install && npm run test:cdk   # testes da stack CDK (TS)
```

Ou abra no GitHub Codespaces — o devcontainer já configura Go, Node, Python, Java, AWS CLI, uv e DynamoDB Local.
