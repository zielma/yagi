# TODO
1. Read API keys from environment.

## Frontend
[ ] select the budget that will be synced & account
[ ] select account from GoCardless to sync with
- [ ] possibility to map budget/account and gocardless account

## Backend

### YNAB
- [ ] fetch budgets
- [ ] fetch the available accounts
- [ ] implement rate limiting

### GoCardless
- [ ] fetch accounts
- [ ] allow users to link bank accounts
- [ ] implement API rate limitation based on [this](https://bankaccountdata.zendesk.com/hc/en-gb/articles/11529584398236-Bank-API-Rate-Limits-and-Rate-Limit-Headers) & [this](https://developer.gocardless.com/bank-account-data/overview)
- [ ]