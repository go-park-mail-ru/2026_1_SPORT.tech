package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	grpcHealth "google.golang.org/grpc/health/grpc_health_v1"
)

type Dependency struct {
	Name     string `json:"name"`
	Endpoint string `json:"endpoint"`
}

type dependencyStatus struct {
	Name     string `json:"name"`
	Endpoint string `json:"endpoint"`
	Status   string `json:"status"`
	Error    string `json:"error,omitempty"`
}

type response struct {
	Status       string             `json:"status"`
	Service      string             `json:"service"`
	Dependencies []dependencyStatus `json:"dependencies"`
}

type GRPCChecker struct {
	dependencies []Dependency
}

func NewGRPCChecker(dependencies []Dependency) *GRPCChecker {
	return &GRPCChecker{dependencies: dependencies}
}

func (checker *GRPCChecker) Check(ctx context.Context) ([]dependencyStatus, bool) {
	statuses := make([]dependencyStatus, 0, len(checker.dependencies))
	isHealthy := true

	for _, dependency := range checker.dependencies {
		status := dependencyStatus{
			Name:     dependency.Name,
			Endpoint: dependency.Endpoint,
			Status:   "ok",
		}

		connection, err := grpc.DialContext(
			ctx,
			dependency.Endpoint,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
		)
		if err != nil {
			status.Status = "degraded"
			status.Error = err.Error()
			statuses = append(statuses, status)
			isHealthy = false
			continue
		}

		client := grpcHealth.NewHealthClient(connection)
		healthResponse, err := client.Check(ctx, &grpcHealth.HealthCheckRequest{})
		_ = connection.Close()
		if err != nil {
			status.Status = "degraded"
			status.Error = err.Error()
			statuses = append(statuses, status)
			isHealthy = false
			continue
		}
		if healthResponse.GetStatus() != grpcHealth.HealthCheckResponse_SERVING {
			status.Status = "degraded"
			status.Error = healthResponse.GetStatus().String()
			isHealthy = false
		}

		statuses = append(statuses, status)
	}

	return statuses, isHealthy
}

func NewHandler(serviceName string, checker *GRPCChecker) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		ctx, cancel := context.WithTimeout(request.Context(), 2*time.Second)
		defer cancel()

		dependencies, isHealthy := checker.Check(ctx)
		payload := response{
			Status:       "ok",
			Service:      serviceName,
			Dependencies: dependencies,
		}
		statusCode := http.StatusOK
		if !isHealthy {
			payload.Status = "degraded"
			statusCode = http.StatusServiceUnavailable
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(statusCode)
		_ = json.NewEncoder(writer).Encode(payload)
	})
}
