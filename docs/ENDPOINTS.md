# Tabela de endpoints validados pelo Motor de Qualidade de Dados (MQD)

Abaixo a lista dos endpoints validados pelo Motor de Qualidade de Dados para cada MVP.

## [Dados Cadastrais e Transacionais]

## [Contas 2.0.1]

| Endpoint Name |
|--|
| /accounts/v2/accounts |
| /accounts/v2/accounts/{accountid} |
| /accounts/v2/accounts/{accountid}/balances |
| /accounts/v2/accounts/{accountid}/transactions |
| /accounts/v2/accounts/{accountid}/transactions-current |
| /accounts/v2/accounts/{accountid}/overdraft-limits |

## [Consentimento 2.2.0]

| Endpoint Name |
|--|
| /consents/v2/consents/{consentId}/extends |
| /consents/v2/consents/{consentId} |

## [Cartão de crédito 2.1.0]

| Endpoint Name |
|--|
| /credit-cards-accounts/v2/accounts |
| /credit-cards-accounts/v2/accounts/{creditCardAccountId} |
| /credit-cards-accounts/v2/accounts/{creditCardAccountId}/bills |
| /credit-cards-accounts/v2/accounts/{creditCardAccountId}/bills/{billId}/transactions |
| /credit-cards-accounts/v2/accounts/{creditCardAccountId}/limits |
| /credit-cards-accounts/v2/accounts/{creditCardAccountId}/transactions |
| /credit-cards-accounts/v2/accounts/{creditCardAccountId}/transactions-current |

## [Dados cadastrais 2.0.1]

| Endpoint Name |
|--|
| /customers/v2/personal/identifications |
| /customers/v2/business/qualifications |
| /customers/v2/business/financial-relations |
| /customers/v2/personal/financial-relations |
| /customers/v2/business/identifications |
| /customers/v2/personal/qualifications |

## [Operações de Crédito - Financiamento 2.1.0]

| Endpoint Name |
|--|
| /financings/v2/contracts/{contractId}/scheduled-instalments |
| /financings/v2/contracts/{contractId}/payments |
| /financings/v2/contracts |
| /financings/v2/contracts/{contractId} |
| /financings/v2/contracts/{contractId}/warranties |

## [Operações de Crédito - Direitos Creditórios Descontados 2.1.0]

| Endpoint Name |
|--|
| /invoice-financings/v2/contracts/{contractId}/warranties |
| /invoice-financings/v2/contracts/{contractId}/payments |
| /invoice-financings/v2/contracts/{contractId} |
| /invoice-financings/v2/contracts/{contractId}/scheduled-instalments |
| /invoice-financings/v2/contracts |

## [Operações de Crédito - Empréstimos 2.1.0]

| Endpoint Name |
|--|
| /loans/v2/contracts |
| /loans/v2/contracts/{contractId} |
| /loans/v2/contracts/{contractId}/warranties |
| /loans/v2/contracts/{contractId}/scheduled-instalments |
| /loans/v2/contracts/{contractId}/payments |

## [Recursos 2.1.0]

| Endpoint Name |
|--|
| /resources/v2/resources |

## [Operações de Crédito - Adiantamento a Depositantes 2.1.0]

| Endpoint Name |
|--|
| /unarranged-accounts-overdraft/v2/contracts/{contractId}/scheduled-instalments |
| /unarranged-accounts-overdraft/v2/contracts/{contractId} |
| /unarranged-accounts-overdraft/v2/contracts |
| /unarranged-accounts-overdraft/v2/contracts/{contractId}/warranties |
| /unarranged-accounts-overdraft/v2/contracts/{contractId}/payments |