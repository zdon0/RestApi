# RestApi
### Realise specifications from [openapi.yaml](https://github.com/zdon0/RestApi/blob/master/openapi.yaml)
Namely CRUD operations for list of products (online market)
- **POST** /imports : Import products with json data
- **DELETE** /delete/id : Delete product and product's children
- **GET** /nodes/id : Present a product's list including children and average price of all product's items (not categories) in json
- **GET** /sales : Present list of items which changed price during specified day in json

### Libraries
[gin](https://github.com/gin-gonic/gin) for route operations and validate parameters <br/>
[pgx](https://github.com/jackc/pgx) for interaction with PostgreSQL

### Tables
#### item
| id, PrimaryKey |   "parentId"   |   name  | price |         type          |    time   |
| -- | -------------- | ------- | ----- | --------------------- | --------- |
|uuid| uuid, nullable | varchar |  int  | enum(OFFER, CATEGORY) | timestamp |
#### price_history
| id, ForeignKey to [item.id](#item) | price | time |
| ---- | --- | --------- |
| uuid | int | timestamp |

### How to use
Install docker and docker-compose  <br/>
Set environment variables USER, PASSWORD, PORT for set postgres user, postgres password and listening port (inside container port is still 8080, so in logs you can see that server starts on 8080) <br/>
Set variables with command ```export PORT=8888``` f.e. <br/>
Then start ``docker-compose up -d`` <br/>
Docker will create volume and fetch dependencies
### To do
- Get /nodes/id/statistic
  May be will realise in the future, if I will have enough time
- Migrate database (probably with python alembic)
- ~~Wrap into docker container~~
- Tune self code when will have more experience
