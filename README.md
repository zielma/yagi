# YAGI
![build](https://github.com/zielma/yagi/actions/workflows/ci.yml/badge.svg) [![codecov](https://codecov.io/gh/zielma/yagi/branch/main/graph/badge.svg)](https://codecov.io/gh/zielma/yagi)

# TODO

## Frontend
- [ ] Bank selection
- [ ] Account mapping YNAB <-> GoCardless
- [ ] Review current mappings
- [ ] Review job run times and schedule

## Backend
- [x] Read API keys from environment.
- [ ] Persistent job schedule in database. 

### YNAB
- [ ] Fetch budgets
- [ ] Fetch the accounts
- [ ] Fetch transactions
- [ ] Handle custom transaction fields mappings for banks
- [ ] Handle rate limitation 

### GoCardless
- [ ] Fetch accounts
- [ ] Handle API rate limitation based on [this](https://developer.gocardless.com/bank-account-data/overview)

### Sequence Diagrams

### <p align="center">User link their bank account(s) with YNAB account(s)</p>
::: mermaid
sequenceDiagram
    participant Bank
    actor User
    autonumber
    User->>YAGI: /binding/add
    activate User
    YAGI-->>User: show bank choice UI
    User->>YAGI: /binding/add?bank_id=BANK_ID
    YAGI->>GoCardless: create requisition /api/v2/requisitions
    GoCardless-->>YAGI: response with redirect url
    YAGI->>User: redirect to redirect url
    User->>+Bank: authorization flow
    Bank->>-User: redirect to YAGI with code
    User->>+YAGI: /binding/add?code=1234567890
    YAGI->>+GoCardless: list accounts for mapping /api/v2/requisitions/{id}
    GoCardless-->>-YAGI: list of accounts
    YAGI->>+YNAB: fetch accounts from YNAB
    YNAB-->>-YAGI: list of accounts
    YAGI-->>-User: show account mapping UI
    note right of User: Show the accounts from YNAB and the accounts from GoCardless <br/>allowing user to map freely which bank account they want to map<br/>to their YNAB account(s). Make it possible to map multiple accounts at once.
    User->>+YAGI: /binding/add
    YAGI-->>-User: show mapped accounts 
    deactivate User
    YAGI->YAGI: store mapping
    YAGI->YAGI: schedule job to fetch transactions
:::
