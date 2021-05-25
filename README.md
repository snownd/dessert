
A dessert for [GIN](https://github.com/gin-gonic/gin) with [fx](https://github.com/uber-go/fx).


## Installation

```bash
go get -u github.com/duanxuelin/dessert
```


## Quick start

1. create a database connection:

```go
type DBConn struct {
  db *sql.DB
}
func NewDBConn(lc fx.Lifecycle) (*DBConn, error){
  // Connect to database
  // Don't forget to close the connection.
  // For example:
  //  lc.Append(fx.Hook{
  // 		OnStop: func(c context.Context) error {
  // 			return db.Close()
  // 		},
  // 	})
}

var DBModule = fx.Provide(NewDBConn)
```

2. create service:

```go
type Foo struct {
  // fieds...
}
type FooService struct {
  dbConn *DBConn
}
// NewFooService dbConn is injected by fx
func NewFooService(dbConn *DBConn) *FooService {
  return &FooService{dbConn}
}

func (s FooService) FindFoo(ctx context.Context, id string) (Foo, error) {
  // find implementation
}

var	ServiceModule := fx.Options(
	fx.Provide(NewPrintService),
  // You can register other services here.
)
```

3. Controller:

```go

type pathParam struct {
	dessert.IPathParam
	ID string `uri:"id" binding:"required,uuid"`
}

type res struct {
	*dessert.JsonBaseResponse
	Foo Foo `json:"foo"`
}

func NewFooController(s *FooService) dessert.Controller {
  ctr := dessert.NewBaseController("/foo")

	ctr.Get("/:id", func(ctx context.Context, pathID *pathParam) (int, dessert.IResponse, error) {
    foo,err := s.FindFoo(ctx, pathID.ID)
		return 200, &res{dessert.Res, foo}, err
	})
}

var ControllerModule = fx.Options(
  // must wrap controller constructor with NewControllerContainer function.
  dessert.NewControllerContainer(NewFooController),
  // Add other controllers here.
)

```

4. App:

```go
func main() {
  app := fx.New(
    DBModule,
    ServiceModule,
    ControllerModule,
    dessert.DefaultServerModule,
  )
  app.Run()
}
```

For runnable code, see [example](./example/main.go).
