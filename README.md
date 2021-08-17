# GQL 
**GQL is small query builder in golang for sql databases in the form of ORM without ORM's helpers,extensions,migrations,complications and ORM's problems.**

**So if you are looking for a way to deal with databases in an organized and easy way while maintaining performance without complications you can use GQL**

## Installation
`go get -u github.com/es-code/gql
`

## Usage

### Create Model
```
func Service() *gql.Model {
	return &gql.Model{Table: "services",Scanner: func() interface{} {
		return &ServiceScanner{}
	}}
}
```
`Scanner : It type contains fields that represent the columns of the table in the database with their data types and their names`
#### Scanner Example :
```
type ServiceScanner struct {
	Id int `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
	Price float32 `db:"price" json:"price"`
}
```

### Selection:

```
services,err:= Service().Get()
```
`sql output: select * from services
`

```
//convert to json directly
Json, _ := json.Marshal(services)
fmt.Println(string(Json))
//json output : [{"id":1,"name":"crm","price":300},{"id":2,"name":"pos","price":100}]

//read data inside loop

for _,service:=range services{
       //cast to our scanner (ServiceScanner) struct to access fields 
        item := service.(*ServiceScanner)
        //price field
	    fmt.Println(item.Price)
	}
```


#### **Where :**


```
services,err:=Service().Where("name","=","sms").Get()
```
`sql output: select * from services where services.name = ?
`
<hr>

#### **Select specific columns :**


```
services,err:=Service().Select("name","price").Where("name","=","pos").Get()
```
`sql output: select name,price from services where services.name = ?
`
<hr>

#### **Order By :**


```
services,err:=Service().OrderBy("id","desc").Get()
```
`sql output: select * from services order by id desc
`
<hr>

#### **Group By :**
```
services,err:=Service().Select("name","count(price) as price").GroupBy("name").OrderBy("price","desc").Get()
```
`sql output: select name,count(price) as price from services group by name order by price desc
`
<hr>

#### **Limit :**
```
services,err:=Service().OrderBy("id","desc").Limit(1).Get()
```
`sql output: select * from services order by id desc limit ?
`
<hr>

#### **First :**
```
service,err:=Service().First()
```
`sql output: select * from services order by id asc limit ?
`

```
service,err:=Service().Where("name","=","crm").First()
```
`sql output: select * from services where services.name = ? order by id asc limit ?
`
<hr>

#### **Find :**

```
service,err:=Service().Find(1)
```
`sql output: select * from services where id = ? limit ?
`
<hr>

#### **Exists :**

```
service,err:=Service().Where("name","=","pos").Exists()
```
`sql output: select exists(select * from services where services.name = ?) as result
`
<hr>

#### **OrWhere :**

```
services,err:=Service().Where("name","=","crm").OrWhere("id","=","1").Get()
```
`sql output: select * from services where services.name = ? or services.id = ?
`

<hr>

#### **Combination Where :**

```
services,err:=Service().Where("name","=","crm").WhereCombination(func(m *gql.Model) {
		 m.Where("price","=","100").OrWhere("price","=","300")
	}).Get()
```
`sql output: select * from services where services.name = ? and (services.price = ? or services.price = ?)
`
<hr>

#### **Where Exists :**
```
services,err:=Service().Where("name","=","crm").WhereExists(func() *gql.Model {
		return models.Client().Where("id","=","1")
	}).Get()
```
`sql output: select * from services where services.name = ? and  exists (select * from clients where clients.id = ?)
`
<hr>

### Insertion:

```
//create struct contains fields that represent the columns of the table and values

service := ServiceScanner{
		Name:  "test insert",
		Price: 300,
	}
	
insertId,err:=Service().Insert(&service)
```
`sql output: INSERT INTO services (name,price) VALUES (?,?)`
<hr>

#### **Insert and return object:**

```
service := ServiceScanner{
		Name:  "test insert",
		Price: 300,
	}
	
service,err:=Service().InsertAndReturn(&service)
```
```
sql output: 
 1 - INSERT INTO services (name,price) VALUES (?,?)
 2 - select * from services where id = ? limit ?
```

<hr>

### Update:

```
//create struct contains fields that represent the columns of the table and values

service := ServiceScanner{
		Name:  "test update",
		Price: 300,
	}
	
affectedRows,err:=Service().Where("id","=","1").Update(&service)
```
`sql output: UPDATE services set name = ?,price = ? where services.id = ?`

<hr>

#### **Update and return object:**

```
service := ServiceScanner{
		Name:  "test update and return updated object",
		Price: 300,
	}
	
updatedServices,err:=Service().Where("id","=","1").UpdateAndReturn(&service)
```
```
sql output:
    1- UPDATE services set name = ?,price = ? where services.id = ?
    2- select * from services where services.id = ?
```
<hr>

### UNION
```
services,_:=models.Service().Union(func() *gql.Model {
	return models.Service().Where("id","=","1")
}).Get()
```
`sql output: select * from services UNION (select * from services where services.id = ?)`
<hr>

### Custom  Scanner
You can create new scanner and use it in your query
```
//Define new struct to use it as scanner
type Scanner struct {
			 Id int `db:"id"`
			 Name string `db:"name"`
		}
//use this struct in query 
services,_:=Service().With("clients").UseScanner(func() interface{} {
		return &Scanner{}
		}).Get()
```
<hr>

### Query Context
```
 services,_:=models.Service().Context(&ctx).Get()
```
<hr>

### Relationships
GQL is using join queries with relationships, so it's solve n+1 problem

#### Define relationship at model
```
func Service() *gql.Model {
    //create model
	model:= &gql.Model{Table: "services",Scanner: func() interface{} {
		return &ServiceScanner{}
	}}
	
	//define this model has relationship with clients table
	model.HasRelation("clients","clients","client_id","id")
	
	// func HasRelation 
	// (relationName string,relatedTable string,foreignKey string,localKey string)
	// you can use HasRelation func to define one-to-one relationship or one-to-many relationship
	return &model
}
```
<hr>

#### Define Many To Many

```
func Service() *gql.Model {
    //create model
	model:= &gql.Model{Table: "services",Scanner: func() interface{} {
		return &ServiceScanner{}
	}}
	
	//define this model has many to many relationship with client table
	model.BelongsToMany("clients","clients","service_id","id","client_id","id","clients_services")

	// func BelongsToMany 
	// (relationName string,relatedTable string,foreignKey string,localKey string,relatedForeignKey string,relatedLocalKey string,middleTable string)
	// you can use BelongsToMany func to define many-to-many relationships
	return &model
}
```
<hr>

#### Select With Relationship

```
services,_:=Service().Select("services.id","clients.name").Where("id","=","1").With("clients").Get()
```
`sql output: select services.id,clients.name from services inner join clients_services on services.id = clients_services.service_id inner join clients on clients.id = clients_services.client_id where services.id = ?
`
##### Note :
    we selected id and name this columns values will be scaned at service scanner and this scanner already has name,id fields,
    so if we select * we must create new struct has fields that represent all output columns , this fields ordered by columns and finaly use this new struct as custom scanner.
<hr>

### Transactions
With GQL you can work with transactions easily and smoothly
```
    //start transaction
	err:= gql.Transaction(&ctx,&sql.TxOptions{}, func(tx *sql.Tx) error {
		
		//select and lock and parse transaction to query we want run in a transaction
		item,err:=models.Service().Where("id","=","1").Transaction(tx).LockForUpdate().First()
		//updated object
		newItem:=&types.Service{Name:"update inside transaction"}
		
		//update and parse transaction to update query
		updatedItem,err=models.Service().Where("id","=","1").Transaction(tx).UpdateAndReturn(newItem)
		return err
	})

	if err != nil{
		log.Println("transaction error:",err)
	}

```
<hr>

### Sql database
You can use a database handler to execute your queries without GQL
```
rows,err:=gql.GetSqlConnection().Query("select * from services")
```























