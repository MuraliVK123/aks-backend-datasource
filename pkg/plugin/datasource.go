package plugin

import (
	
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	// "crypto/tls"
	"strconv"
	"time"
	"log"
	"os"
	//"io"
	"crypto/tls"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/httpclient"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	// "github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/backend/tracing"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Make sure Datasource implements required interfaces. This is important to do
// since otherwise we will only get a not implemented error response from plugin in
// runtime. In this example datasource instance implements backend.QueryDataHandler,
// backend.CheckHealthHandler interfaces. Plugin should not implement all these
// interfaces- only those which are required for a particular task.
var (
	_ backend.QueryDataHandler      = (*Datasource)(nil)
	_ backend.CheckHealthHandler    = (*Datasource)(nil)
	_ instancemgmt.InstanceDisposer = (*Datasource)(nil)
)

var (
	errRemoteRequest  = errors.New("remote request error")
	errRemoteResponse = errors.New("remote response error")
)

// // NewDatasource creates a new datasource instance.
// func NewDatasource(_ backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
// 	return &Datasource{}, nil
// }

func NewDatasource(settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	opts, err := settings.HTTPClientOptions()
	if err != nil {
		return nil, fmt.Errorf("http client options: %w", err)
	}
	// Using httpclient.New without any provided httpclient.Options creates a new HTTP client with a set of
	// default middlewares (httpclient.DefaultMiddlewares) providing additional built-in functionality, such as:
	//	- TracingMiddleware (creates spans for each outgoing HTTP request)
	//	- BasicAuthenticationMiddleware (populates Authorization header if basic authentication been configured via the
	//		DataSourceHttpSettings component from @grafana/ui)
	//	- CustomHeadersMiddleware (populates headers if Custom HTTP Headers been configured via the DataSourceHttpSettings
	//		component from @grafana/ui)
	//	- ContextualMiddleware (custom middlewares per context.Context, e.g. forwarding HTTP headers based on Allowed cookies
	//		and Forward OAuth Identity configured via the DataSourceHttpSettings component from @grafana/ui)
	cl, err := httpclient.New(opts)
	if err != nil {
		return nil, fmt.Errorf("httpclient new: %w", err)
	}
	return &Datasource{
		settings:   settings,
		httpClient: cl,
	}, nil
}

// DatasourceOpts contains the default ManageOpts for the datasource.
var DatasourceOpts = datasource.ManageOpts{
	TracingOpts: tracing.Opts{
		// Optional custom attributes attached to the tracer's resource.
		// The tracer will already have some SDK and runtime ones pre-populated.
		CustomAttributes: []attribute.KeyValue{
			attribute.String("my_plugin.my_attribute", "custom value"),
		},
	},
}

// Datasource is an example datasource which can respond to data queries, reports
// its health and has streaming skills.
type Datasource struct{
	settings backend.DataSourceInstanceSettings

	httpClient *http.Client


}

// Dispose here tells plugin SDK that plugin wants to clean up resources when a new instance
// created. As soon as datasource settings change detected by SDK old datasource instance will
// be disposed and a new one will be created using NewSampleDatasource factory function.
func (d *Datasource) Dispose() {
	// Clean up datasource instance resources.
	d.httpClient.CloseIdleConnections()
}

var (
    WarningLog *log.Logger
    InfoLog   *log.Logger
    ErrorLog   *log.Logger
)
func init() {
    file, err := os.OpenFile("myLOG.txt", os.O_RDWR | os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
    if err != nil {
        log.Fatal(err)
    }

    InfoLog = log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
    WarningLog = log.New(file, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
    ErrorLog = log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifier).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (d *Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {

	response := backend.NewQueryDataResponse()
	
	fmt.Sprintf("user friendly error for query number  , excluding any sensitive information")

	// loop over queries and execute them individually.
	for i, q := range req.Queries {
		if i%2 != 0 {
			// Just to demonstrate how to return an error with a custom status code.
			response.Responses[q.RefID] = backend.ErrDataResponse(
				backend.StatusBadRequest,
				fmt.Sprintf("user friendly error for query number %v, excluding any sensitive information", i+1),
			)
			continue
		}

		res, err := d.query(ctx, req.PluginContext, q)
		switch {
		case err == nil:
			break
		case errors.Is(err, context.DeadlineExceeded):
			res = backend.ErrDataResponse(backend.StatusTimeout, "gateway timeout")
		case errors.Is(err, errRemoteRequest):
			res = backend.ErrDataResponse(backend.StatusBadGateway, err.Error())
		case errors.Is(err, errRemoteResponse):
			res = backend.ErrDataResponse(backend.StatusValidationFailed, err.Error())
		default:
			res = backend.ErrDataResponse(backend.StatusInternal, err.Error())
		}
		// save the response in a hashmap
		// based on with RefID as identifier
		response.Responses[q.RefID] = res
		log.Println(res)
	}
	fmt.Println(response)
	return response, nil
}


func parseJSON[T apiMetrics] (s []byte) (apiMetrics) {
	var r apiMetrics
	if err := json.Unmarshal(s, &r); err != nil {
	  return r
	}
	return r
  }

func (d *Datasource) query(ctx context.Context, pCtx backend.PluginContext, query backend.DataQuery) (backend.DataResponse, error) {
	// Create spans for this function.
	// tracing.DefaultTracer() returns the tracer initialized when calling Manage().
	// Refer to OpenTelemetry's Go SDK to know how to customize your spans.
	// InfoLog.Println(ctx)
	// InfoLog.Println(query)

	ctx, span := tracing.DefaultTracer().Start(
		ctx,
		"query processing",
		trace.WithAttributes(
			attribute.String("query.ref_id", query.RefID),
			attribute.String("query.type", query.QueryType),
			attribute.Int64("query.max_data_points", query.MaxDataPoints),
			attribute.Int64("query.interval_ms", query.Interval.Milliseconds()),
			attribute.Int64("query.time_range.from", query.TimeRange.From.Unix()),
			attribute.Int64("query.time_range.to", query.TimeRange.To.Unix()),
		),
	)
	defer span.End()
    //url = query.target
	var input interface{}
	var signal string;
	var targetURL string;
	var aliasName string;
	var scale float64 = 1;
    
	if len(query.JSON) > 0 {
		err := json.Unmarshal(query.JSON, &input)
		targetURL = d.settings.URL + "/api/realtime/"
		switch data:= input.(type) {
		case map[string]interface{}:
			targetURL = targetURL + data["type"].(string) + "/?db=global" 
			signal = data["target"].(string);
			if item,ok := data["alias"].(string); ok{
				aliasName = item
			}else{
				aliasName = signal
			}
			if mul,ok := data["scale"].(string); ok{
				scale,err = strconv.ParseFloat(mul, 64)
				if(scale == 0){
					scale = 1
				}
			}else{
				scale = 1
			}
			
			
			targetURL = targetURL + "&signal=" + signal;
			InfoLog.Println(err)

		}
		InfoLog.Println(input,err)

	}

	if signal == "" {
		return backend.DataResponse{},nil
	}

 
	//return backend.DataResponse{}, fmt.Errorf("%w: expected 200 response, got %d", username,"scale")


	// Do HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, string(targetURL), nil)	
	if err != nil {
		return backend.DataResponse{}, fmt.Errorf("new request with context: %w", err)
	}

	if password,exists  := d.settings.DecryptedSecureJSONData["basicAuthPassword"] ; exists{
		req.SetBasicAuth(d.settings.BasicAuthUser, password)
	}
	
	
	if len(query.JSON) > 0 {
		
		if err != nil {
			return backend.DataResponse{}, fmt.Errorf("unmarshal: %w", err)
		}
		q := req.URL.Query()
		req.URL.RawQuery = q.Encode()
	}

	

	InfoLog.Println(req.URL.RawQuery)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}


	InfoLog.Println(req.URL.RawQuery)
	client := &http.Client{Transport: tr}
	httpResp, err := client.Do(req)	
	switch {
	case err == nil:
		break
	case errors.Is(err, context.DeadlineExceeded):
		return backend.DataResponse{}, err
	default:
		return backend.DataResponse{}, fmt.Errorf("http client do: %w: %s", errRemoteRequest, err)
	}
	defer func() {
		if err := httpResp.Body.Close(); err != nil {
			// log.DefaultLogger.Error("query: failed to close response body", "err", err)
		}
	}()
	//return backend.DataResponse{}, fmt.Errorf("%w: expected 200 response, got %d", req.URL, httpResp.StatusCode)

	span.AddEvent("HTTP request done")
	InfoLog.Println("HTTP request done")
	// Make sure the response was successful
	if httpResp.StatusCode != http.StatusOK {
		return backend.DataResponse{}, fmt.Errorf("%w: expected 200 response, got %d", errRemoteResponse, httpResp.StatusCode)
	}
	
	defer httpResp.Body.Close()

	if err != nil {
	  return backend.DataResponse{}, err
	}

	
	var responseData interface{}
	err = json.NewDecoder(httpResp.Body).Decode(&responseData)
	var timestamps string
	var times []time.Time
	var values []float64

    
	switch data:= responseData.(type) {
	case map[string]interface{}:
		result := data[signal].(map[string]interface{})
		valueCount := result["value"].(float64)
		values = append(values,valueCount * scale)
		timestamps = result["timestamp"].(string)
		date,error := time.Parse(time.RFC3339, timestamps)
		fmt.Println(error)
		times = append(times,date)
	case []interface{}:
		for _, item := range data {
			if itemMap, ok := item.(map[string]interface{}); ok {
			   valueCount := itemMap[signal].(float64)
			   values = append(values,valueCount * scale)
			   logtimestamps := itemMap["timestamp"].(float64)
			   t := time.Unix(int64(logtimestamps) / 1000,0)
			   times = append(times,t)
			}
		}
		
	}

	if err != nil {
	  return backend.DataResponse{}, err
	}
	//body = parseJSON(bodyJ)

	span.AddEvent("JSON response decoded")
	// InfoLog.Println(body.DataPoints)

	// Create slice of values for time and values.
	dataResp := backend.DataResponse{
		Frames: []*data.Frame{
			data.NewFrame(
				aliasName,
				data.NewField("time", nil, times),
				data.NewField(signal, nil, values),
			),
		},
	}
	
	span.AddEvent("Frames created")
	InfoLog.Println("Frames created")

	return dataResp, err
}



// CheckHealth performs a request to the specified data source and returns an error if the HTTP handler did not return
// a 200 OK response.
func (d *Datasource) CheckHealth(ctx context.Context, _ *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, d.settings.URL, nil)
	if err != nil {
		return newHealthCheckErrorf("could not create request"), nil
	}
    

	tr := http.DefaultTransport.(*http.Transport)
	tr.TLSClientConfig.InsecureSkipVerify = true
    client := &http.Client{Transport: tr}
    resp, err := client.Get(d.settings.URL)
    if err != nil {
        fmt.Println(err)
	    InfoLog.Println(err)

	}
	


	resp1, err := d.httpClient.Do(r)
	// if err != nil {
	// 	return newHealthCheckErrorf("sew request error", err), nil
	// }
	defer func() {
		if err := resp1.Body.Close(); err != nil {
			// log.DefaultLogger.Error("check health: failed to close response body", "err", err.Error())
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return newHealthCheckErrorf("got response code %d", resp.StatusCode), nil
	}
	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "Data source is working",
	}, nil
}

// newHealthCheckErrorf returns a new *backend.CheckHealthResult with its status set to backend.HealthStatusError
// and the specified message, which is formatted with Sprintf.
func newHealthCheckErrorf(format string, args ...interface{}) *backend.CheckHealthResult {
	return &backend.CheckHealthResult{Status: backend.HealthStatusError, Message: fmt.Sprintf(format, args...)}
}
