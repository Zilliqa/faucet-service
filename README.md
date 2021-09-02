# faucet-service

## Prerequisites

- [Golang](https://golang.org)
- [Docker](https://www.docker.com/community-edition)

## Environment Variables

Create a `.env` file with the following variables

| Name               | Description                                              | Type         | Examples                                      |
| ------------------ | -------------------------------------------------------- | ------------ | --------------------------------------------- |
| `ENV_TYPE`         | Environment Type                                         | String       | `dev`                                         |
| `NODE_URL`         | Node URL                                                 | String       | `https://zilliqa-isolated-server.zilliqa.com` |
| `CHAIN_ID`         | Chain ID                                                 | String       | `222`                                         |
| `AMOUNT_IN_ZIL`    | Fund amount in ZIL                                       | String       | `1000`                                        |
| `BATCH_INTERVAL`   | Batch interval. Valid time units are `ms`, `s`, `m`, `h` | String       | `15s`, `5m`                                   |
| `BATCH_LIMIT`      | Batch limit                                              | String       | `1000`                                        |
| `TTL`              | Time To Live                                             | String       | `300`                                         |
| `PRIVATE_KEY`      | Private key of the account to be used for funding        | SecureString |                                               |
| `RECAPTCHA_SECRET` | reCAPTCHA secret                                         | SecureString |                                               |

## Installation and Usage

### `make deps`

Installs dependencies.

### `make build`

Builds the project.

### `make test`

Runs tests.

### `make cover`

Shows an HTML presentation of the source code decorated with coverage information.

### `make start`

Runs Docker container.

## API Documentation

View documentation on <a>https://editor.swagger.io/</a> by pasting openapi.yml from root directory (Recommended)
Or, install vscode plugins for openapi to be able to preview it using swaggerUI.

## Database

This service uses [go-memdb](https://github.com/hashicorp/go-memdb) which is a simple in-memory database built on [immutable radix trees](https://github.com/hashicorp/go-immutable-radix).

## Workflow Visualization

### Cron job

The cron job runs the following 4 functions:

- `Confirm()`
- `Expire()`
- `Retry()`
- `Send()`

The following table contains sample items and we will see how the above 4 functions change the table.

| Status | ID        | CreatedAt              | Address     | TxID      |
| ------ | --------- | ---------------------- | ----------- | --------- |
| ⌛️    | `...a4a0` | `2021-01-01T00:00:07Z` | `0x...9e79` | `...8e10` |
| ✅     | `...a4a1` | `2021-01-01T00:16:07Z` | `0x...9e79` | `...8e11` |
| ✅     | `...a4a2` | `2021-01-01T00:17:07Z` | `0x...9e79` | `...8e12` |
| ✅     | `...a4a3` | `2021-01-01T00:17:07Z` | `0x...9e79` | `...8e13` |
| 👁      | `...a4a4` | `2021-01-01T00:18:07Z` | `0x...9e79` | `...8e14` |
| 👁      | `...a4a5` | `2021-01-01T00:19:07Z` | `0x...9e79` | `...8e15` |
| 📦     | `...a4a6` | `2021-01-01T00:25:07Z` | `0x...9e79` |           |
| 📦     | `...a4a7` | `2021-01-01T00:26:07Z` | `0x...9e79` |           |

Let's assume the following:

- batch interval is `5 min`.
- estimated confirmation time is `3 min`.
- The batch interval is always longer than the estimated confirmation time.
- Time To Live is `30 min`.
- Now is `2021-01-01T00:31:07Z`.
- It's OK to remove the expired messages.
- It's OK to process a message more than once.

Also, note that each item has one of the following status:

- ✅ Confirmed
- ⌛️ Expired
- 👁 Unconfirmed (being tracked)
- 📦 Pending

#### 1. `Confirm()` Deletes the confirmed items which are no longer needed.

| Status | ID        | CreatedAt              | Address     | TxID      |
| ------ | --------- | ---------------------- | ----------- | --------- |
| ⌛️    | `...a4a0` | `2021-01-01T00:00:07Z` | `0x...9e79` | `...8e10` |
| ✅     | `...a4a1` | `2021-01-01T00:16:07Z` | `0x...9e79` | `...8e11` |
| ✅     | `...a4a2` | `2021-01-01T00:17:07Z` | `0x...9e79` | `...8e12` |
| ✅     | `...a4a3` | `2021-01-01T00:17:07Z` | `0x...9e79` | `...8e13` |
| 👁      | `...a4a4` | `2021-01-01T00:18:07Z` | `0x...9e79` | `...8e14` |
| 👁      | `...a4a5` | `2021-01-01T00:19:07Z` | `0x...9e79` | `...8e15` |
| 📦     | `...a4a6` | `2021-01-01T00:25:07Z` | `0x...9e79` |           |
| 📦     | `...a4a7` | `2021-01-01T00:26:07Z` | `0x...9e79` |           |

After `Confirm()` the table will be the following:

| Status | ID        | CreatedAt              | Address     | TxID      |
| ------ | --------- | ---------------------- | ----------- | --------- |
| ⌛️    | `...a4a0` | `2021-01-01T00:00:07Z` | `0x...9e79` | `...8e10` |
| 👁      | `...a4a4` | `2021-01-01T00:18:07Z` | `0x...9e79` | `...8e14` |
| 👁      | `...a4a5` | `2021-01-01T00:19:07Z` | `0x...9e79` | `...8e15` |
| 📦     | `...a4a6` | `2021-01-01T00:25:07Z` | `0x...9e79` |           |
| 📦     | `...a4a7` | `2021-01-01T00:26:07Z` | `0x...9e79` |           |

Note that the items with ✅ status have been deleted.

#### 2. `Expire()` Reduces stored data volumes by expiring the old items.

| Status | ID        | CreatedAt              | Address     | TxID      |
| :----: | --------- | ---------------------- | ----------- | --------- |
|  ⌛️   | `...a4a0` | `2021-01-01T00:00:07Z` | `0x...9e79` | `...8e10` |
|   👁    | `...a4a4` | `2021-01-01T00:18:07Z` | `0x...9e79` | `...8e14` |
|   👁    | `...a4a5` | `2021-01-01T00:19:07Z` | `0x...9e79` | `...8e15` |
|   📦   | `...a4a6` | `2021-01-01T00:25:07Z` | `0x...9e79` |           |
|   📦   | `...a4a7` | `2021-01-01T00:26:07Z` | `0x...9e79` |           |

For some reason `...a4a0` is still not confirmed and it should be expired because of the following reasons:

- `...a4a0` was created at `2021-01-01T00:00:07Z`
- Time To Live is `30 min`
- now is `2021-01-01T00:31:07Z`

After `Expire()` the table will be the following:

| Status | ID        | CreatedAt              | Address     | TxID      |
| :----: | --------- | ---------------------- | ----------- | --------- |
|   👁    | `...a4a4` | `2021-01-01T00:18:07Z` | `0x...9e79` | `...8e14` |
|   👁    | `...a4a5` | `2021-01-01T00:19:07Z` | `0x...9e79` | `...8e15` |
|   📦   | `...a4a6` | `2021-01-01T00:25:07Z` | `0x...9e79` |           |
|   📦   | `...a4a7` | `2021-01-01T00:26:07Z` | `0x...9e79` |           |

Note that the items with ⌛️ status have been deleted.

#### 3. `Retry()` Removes the old tx id.

| Status | ID        | CreatedAt              | Address     | TxID      |
| :----: | --------- | ---------------------- | ----------- | --------- |
|   👁    | `...a4a4` | `2021-01-01T00:18:07Z` | `0x...9e79` | `...8e14` |
|   👁    | `...a4a5` | `2021-01-01T00:19:07Z` | `0x...9e79` | `...8e15` |
|   📦   | `...a4a6` | `2021-01-01T00:25:07Z` | `0x...9e79` |           |
|   📦   | `...a4a7` | `2021-01-01T00:26:07Z` | `0x...9e79` |           |

Note that it's at-least-once delivery.

After `Retry()` the table will be the following:

| Status | ID        | CreatedAt              | Address     | TxID |
| :----: | --------- | ---------------------- | ----------- | ---- |
|   📦   | `...a4a4` | `2021-08-29T02:09:07Z` | `0x...9e79` |      |
|   📦   | `...a4a5` | `2021-08-29T02:10:07Z` | `0x...9e79` |      |
|   📦   | `...a4a6` | `2021-01-01T00:25:07Z` | `0x...9e79` |      |
|   📦   | `...a4a7` | `2021-01-01T00:26:07Z` | `0x...9e79` |      |

Note that the items doesn't have the tx ID.

#### 4. `Send()` Creates transactions.

| Status | ID        | CreatedAt              | Address     | TxID |
| :----: | --------- | ---------------------- | ----------- | ---- |
|   📦   | `...a4a4` | `2021-08-29T02:09:07Z` | `0x...9e79` |      |
|   📦   | `...a4a5` | `2021-08-29T02:10:07Z` | `0x...9e79` |      |
|   📦   | `...a4a6` | `2021-01-01T00:25:07Z` | `0x...9e79` |      |
|   📦   | `...a4a7` | `2021-01-01T00:26:07Z` | `0x...9e79` |      |

After `Send()` the table will be the following:

| Status | ID        | CreatedAt              | Address     | TxID      |
| :----: | --------- | ---------------------- | ----------- | --------- |
|   👁    | `...a4a4` | `2021-08-29T02:09:07Z` | `0x...9e79` | `...0e10` |
|   👁    | `...a4a5` | `2021-08-29T02:10:07Z` | `0x...9e79` | `...0e11` |
|   👁    | `...a4a6` | `2021-01-01T00:25:07Z` | `0x...9e79` | `...0e12` |
|   👁    | `...a4a7` | `2021-01-01T00:26:07Z` | `0x...9e79` | `...0e13` |

Note that each items have the new tx ID.

## Licence

You can view our [licence here](LICENSE).
