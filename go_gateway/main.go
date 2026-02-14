package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/shattang/RiskGo/go_gateway/internal/market"
	pb "github.com/shattang/RiskGo/go_gateway/proto"
)

type AnalyzeRequest struct {
	Positions     []Position `json:"positions"`
	ScenarioRange []float64  `json:"scenario_range"`
	Volatility    float64    `json:"volatility"`
}

type Position struct {
	Ticker   string  `json:"ticker"`
	Quantity float64 `json:"quantity"`
	Beta     float64 `json:"beta"`
	Legs     []Leg   `json:"legs"`
}

type Leg struct {
	Type   string  `json:"type"`   // "CALL" or "PUT"
	Strike float64 `json:"strike"`
	Expiry string  `json:"expiry"` // YYYY-MM-DD
}

type Metrics struct {
	PnL   float64 `json:"pnl"`
	Delta float64 `json:"delta"`
	Gamma float64 `json:"gamma"`
	Theta float64 `json:"theta"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	conn, err := grpc.DialContext(ctx, "cpp_engine:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewRiskEngineClient(conn)

	marketProvider := market.NewYahooFinanceProvider(1 * time.Minute)

	app := fiber.New()

	app.Post("/api/analyze_portfolio", func(c *fiber.Ctx) error {
		var req AnalyzeRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
		}

		type tickerData struct {
			spot float64
			vol  float64
			rfr  float64
		}
		marketCache := make(map[string]tickerData)
		results := make(map[string]map[string]Metrics)

		for _, pos := range req.Positions {
			if _, ok := results[pos.Ticker]; !ok {
				results[pos.Ticker] = make(map[string]Metrics)
			}
			
			data, ok := marketCache[pos.Ticker]
			if !ok {
				spot, err := marketProvider.GetSpotPrice(c.Context(), pos.Ticker)
				if err != nil {
					log.Printf("Error getting spot price for %s: %v", pos.Ticker, err)
					continue
				}
				rfr, _ := marketProvider.GetRiskFreeRate(c.Context())
				
				vol := req.Volatility
				if vol == 0 {
					vol, _ = marketProvider.GetVolatility(c.Context(), pos.Ticker)
				}
				
				data = tickerData{spot, vol, rfr}
				marketCache[pos.Ticker] = data
			}

			for _, shock := range req.ScenarioRange {
				grpcReq := &pb.ScenarioRequest{
					SpotPrice:         data.spot,
					RiskFreeRate:      data.rfr,
					Volatility:        data.vol,
					ScenarioPctChange: shock,
					Beta:              pos.Beta,
					Legs:              make([]*pb.OptionLeg, 0),
				}

				for _, leg := range pos.Legs {
					legType := pb.OptionLeg_CALL
					if leg.Type == "PUT" {
						legType = pb.OptionLeg_PUT
					}
					grpcReq.Legs = append(grpcReq.Legs, &pb.OptionLeg{
						Type:     legType,
						Strike:   leg.Strike,
						Expiry:   leg.Expiry,
						Quantity: pos.Quantity,
					})
				}

				resp, err := client.CalculateBetaScenario(c.Context(), grpcReq)
				if err != nil {
					log.Printf("gRPC error for %s scenario %f: %v", pos.Ticker, shock, err)
				} else {
					shockStr := fmt.Sprintf("%.2f", shock)
					m := results[pos.Ticker][shockStr]
					m.PnL += resp.ScenarioPnl
					m.Delta += resp.ScenarioDelta
					m.Gamma += resp.ScenarioGamma
					m.Theta += resp.ScenarioTheta
					results[pos.Ticker][shockStr] = m
				}
			}
		}

		return c.JSON(results)
	})

	log.Println("Go Gateway listening on :3000")
	log.Fatal(app.Listen(":3000"))
}
