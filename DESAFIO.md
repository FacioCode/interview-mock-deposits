# Desafio: Instant Deposits

Serviço serverless em Go que processa depósitos instantâneos: recebe eventos via EventBridge, persiste em DynamoDB, dispara pagamentos a um parceiro externo e reflete o resultado de volta no depósito.

## Sua tarefa

Complete o ciclo de vida de um depósito:

1. Após a criação, disparar o pagamento no parceiro.
2. Ao receber o webhook de retorno do parceiro, refletir o resultado no estado do depósito (sucesso, falha, devolvido).

Cubra o fluxo com testes. O lambda `create_new_deposit/` está completo e serve como referência (handler + testes de integração + acesso a dados).

## Estrutura

| Componente             | Estado                      |
| ---------------------- | --------------------------- |
| `create_new_deposit/`  | implementado (referência)   |
| `deposit_in_provider/` | incompleto                  |
| `webhook/`             | incompleto                  |
| `partner_api/`         | stub                        |
| `deposit/src/`         | algumas funções como stub   |
| `stream_consumer/`     | DDB stream → eventos        |

## Como rodar

Abra no Codespaces (o devcontainer já configura Go, Node, Java e AWS CLI) ou rode localmente:

```bash
cd go
npm install
npm test            # Go + DynamoDB Local
npm run test:cdk    # CDK
```
