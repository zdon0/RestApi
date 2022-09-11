# RestApi
### Realise specifications from [openapi.yaml](https://github.com/zdon0/RestApi/blob/master/openapi.yaml)
Namely CRUD operations for list of products (online market)
- **POST** /imports : Import products with json data
- **DELETE** /delete/id : Delete product and product's children
- **GET** /nodes/id : Present a product's list including children and average price of all product's items (not categories) in json
- **GET** /sales : Present list of items which changed price during specified day in json

### Libraries
[gin](https://github.com/gin-gonic/gin) for route operations and validate parameters <br />
[pgx](https://github.com/jackc/pgx) for interaction with PostgreSQL

### Tables
#### item
| id, PrimaryKey |   "parentId"   |   name  | price |         type          |    time   |
| -- | -------------- | ------- | ----- | --------------------- | --------- |
|uuid| uuid, nullable | varchar |  int  | enum(OFFER, CATEGORY) | timestamp |
#### price_history
| id, ForeignKey to [item.id](#item) | price | time |
| -- | ----- | ---- |
| uuid | int | timestamp |

### How to build
```
go build -tags=jsoniter .
```
[DropIn replace for std jsonEncoder](https://github.com/gin-gonic/gin#build-with-json-replacement)

### How to use
```sh
./RestApi -port=$1 -user=$2 -password=$3
```
- **-port** determines localhost port
- **-user** and **-password** use for database user and password respectively
### To do
- Get /nodes/id/statistic
  May be will realise in the future, if I will have enough time
- Migrate database (probably with python alembic)
- Wrap into docker container
- Tune self code when will have more experience
