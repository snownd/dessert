package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/snownd/dessert"
	"go.uber.org/fx"
)

type exampleDTO struct {
	dessert.IDTO
	Name string `json:"name" binding:"required,alphanum"`
	Date int64  `json:"date" binding:"required,numeric"`
}

type headerData struct {
	dessert.IHeader
	AccessToken string `header:"Access-Token"`
}

type res struct {
	*dessert.JsonBaseResponse
	Token string
}

type PrintService struct {
}

// NewPrintService  PrintService constructor.
// You can add some dependencies provided by fx, *sql.DB for example.
func NewPrintService() *PrintService {
	return &PrintService{}
}

func (s PrintService) Print(ctx context.Context, data interface{}) {
	ticker := time.NewTicker(1 * time.Second)
	j, _ := json.Marshal(data)
	for {
		select {
		case <-ctx.Done():
			return
		case t := <-ticker.C:
			fmt.Printf("%v printService out: %s\n", t, j)
		}
	}
}

// NewExampleController controller constructor.
// You can have any number of controllers whose dependencies are injected by fx.
// You don't need to concern about how to create or destroy those dependencies.
func NewExampleController(ps *PrintService) dessert.Controller {
	ctr := dessert.NewBaseController("/dessert")

	// curl -v -X PUT -H 'Content-Type:application/json' -d '{"name":"foo", "date": 1609459200000}' localhost:3000/dessert/print
	ctr.Put("/print", func(ctx context.Context, data *exampleDTO) {
		c, cacel := context.WithTimeout(ctx, 3*time.Second)
		defer cacel()
		ps.Print(c, data)
	})

	// curl -H 'Access-Token: a6661260-bc68-11eb-9e9d-1904606b61ec' localhost:3000/dessert/token
	// Handler function can return three values or none.
	// Three values returned are http status code, http response and error.
	ctr.Get("/token", func(h *headerData) (int, dessert.IResponse, error) {
		return 200, &res{dessert.Res, h.AccessToken}, nil
	})

	// curl -X PUT localhost:3000/dessert/panic
	ctr.Put("/panic", func() {
		panic(errors.New("/dessert/panic"))
	})
	return ctr
}

func main() {

	controllerModule := fx.Options(
		fx.Provide(dessert.NewControllerContainer(NewExampleController)),
	)
	serviceModule := fx.Options(
		fx.Provide(NewPrintService),
	)
	serverModule := dessert.DefaultServerModule

	app := fx.New(
		controllerModule,
		serviceModule,
		serverModule,
	)
	app.Run()
}
