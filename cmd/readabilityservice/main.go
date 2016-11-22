package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	restful "github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful/swagger"
	"github.com/the42/adequate/portalwatch"
	"github.com/the42/readability"
)

type ReadabilityRequest struct {
	CheckString     *string `description:"Input String whose readability should be checked"`
	CorrelationID   *string `description:"request provided CorrelationID copied to response for requests/response matchmaking"`
	ReadabilityType *string `description:"Algorithm to use for readability check"`
}

type ReadabilityResponse struct {
	ReadabilityRequest ReadabilityRequest
	Response           struct {
		Readability float32 `description:"Readability score result"`
		Message     *string `description:"diagnostic message returned by readability ccheck"`
		StatusCode  int     `description:"0:success, -1: no success, check Message"`
	}
}
type PortalReadabilityRequest struct {
	CKANMDAustria   *portalwatch.CKANMDAustria `description:"the raw CKAN metadata harvested"`
	CorrelationID   *string                    `description:"request provided CorrelationID copied to response for requests/response matchmaking"`
	ReadabilityType *string                    `description:"Algorithm to use for readability check"`
}

type PortalReadabilityResponse struct {
	PortalReadabilityRequest PortalReadabilityRequest `description:"Copied to response from Request`
	Response                 struct {
		Readability float32 `description:"Readability score result"`
		CheckString *string `description:"The actual tested string"`
		Message     *string `description:"diagnostic message returned by readability ccheck"`
		StatusCode  int     `description:"0:success, -1: no success, check Message"`
	}
}

var readabilityrequesttypemappings = map[string]readability.CompareType{
	"WSTF1": readability.WSTF1,
	"WSTF2": readability.WSTF2,
	"WSTF3": readability.WSTF3,
	"WSTF4": readability.WSTF4,
}

type readabilityservice struct {
	r *readability.Readability
}

func appendclosingiandblank(in string) string {
	if len(in) > 0 {
		last_char := in[len(in)-1]
		if !(last_char == '.' || last_char == '!' || last_char == '?') {
			return in + ". "
		}
		return in + " "
	}
	return in
}

func (s *readabilityservice) portalreadabilityservice(request *restful.Request, response *restful.Response) {

	readabilityrequest := PortalReadabilityRequest{}
	if err := request.ReadEntity(&readabilityrequest); err != nil {
		logresponse(response, http.StatusBadRequest, fmt.Sprintf("unable to parse request: %s", err.Error()))
		return
	}

	if readabilityrequest.CKANMDAustria == nil {
		logresponse(response, http.StatusBadRequest, fmt.Sprintf("PortalReadabilityRequest.CKANMDAustriaportal is required but not set"))
		return
	}

	result := PortalReadabilityResponse{PortalReadabilityRequest: readabilityrequest}
	// set the input struct to nil for performance reasons
	result.PortalReadabilityRequest.CKANMDAustria = nil

	var readability_type readability.CompareType
	if readabilityrequest.ReadabilityType != nil && len(*readabilityrequest.ReadabilityType) > 0 {
		readability_type = readabilityrequesttypemappings[*readabilityrequest.ReadabilityType]
	} else {
		readability_type = readability.WSTF1
	}

	// prepare input data for readability check
	// The algorithm is as follows:
	//   if the notes - field doesn't end with a '.', add one (WSTF operates on the notion of a "sentence")
	//   if the title - field doesn't end with a '.', add one (WSTF operates on the notion of a "sentence")
	// The resulting readability is performed as the WSTF on "notes + title + tags" (in this order).

	var readability_inputstring string
	if readabilityrequest.CKANMDAustria.Notes != nil {
		readability_inputstring = *readabilityrequest.CKANMDAustria.Notes
		readability_inputstring = appendclosingiandblank(readability_inputstring)
	}
	if readabilityrequest.CKANMDAustria.Title != nil {
		readability_inputstring += *readabilityrequest.CKANMDAustria.Title
		readability_inputstring = appendclosingiandblank(readability_inputstring)
	}
	for _, v := range readabilityrequest.CKANMDAustria.Tags {
		if dn := v.DisplayName; dn != nil {
			readability_inputstring += " " + *dn
		}
	}

	result.Response.CheckString = &readability_inputstring

	switch readability_type {
	case readability.WSTF1, readability.WSTF2, readability.WSTF3, readability.WSTF4:
		readabilityresult, err := s.r.WienerSachTextFormelType(readability_inputstring, readability_type)
		if err != nil {
			logresponse(response, http.StatusBadRequest, fmt.Sprintf("WienerSachTextFormelType returned error: %s", err.Error()))
			return
		}
		result.Response.Readability = readabilityresult
	default:
		result.Response.StatusCode = -1
		s := "no method found to perform readability check"
		result.Response.Message = &s
	}
	response.WriteAsJson(result)
}

func (s *readabilityservice) readabilityservice(request *restful.Request, response *restful.Response) {

	readabilityrequest := ReadabilityRequest{}
	if err := request.ReadEntity(&readabilityrequest); err != nil {
		logresponse(response, http.StatusBadRequest, fmt.Sprintf("unable to parse request: %s", err.Error()))
		return
	}
	if readabilityrequest.CheckString == nil {
		logresponse(response, http.StatusBadRequest, fmt.Sprintf("ReadabilitySimpleRequest.CheckString is required but not set"))
		return
	}

	var readability_type readability.CompareType

	if readabilityrequest.ReadabilityType != nil && len(*readabilityrequest.ReadabilityType) > 0 {
		readability_type = readabilityrequesttypemappings[*readabilityrequest.ReadabilityType]
	} else {
		readability_type = readability.WSTF1
	}

	result := ReadabilityResponse{ReadabilityRequest: readabilityrequest}
	// set the input string to nil for performance reasons. May correlate result to request by using CorrelationID
	result.ReadabilityRequest.CheckString = nil

	switch readability_type {
	case readability.WSTF1, readability.WSTF2, readability.WSTF3, readability.WSTF4:
		readabilityresult, err := s.r.WienerSachTextFormelType(*readabilityrequest.CheckString, readability_type)
		if err != nil {
			logresponse(response, http.StatusBadRequest, fmt.Sprintf("WienerSachTextFormelType returned error: %s", err.Error()))
			return
		}
		result.Response.Readability = readabilityresult
	default:
		result.Response.StatusCode = -1
		s := "no method found to perform readability check"
		result.Response.Message = &s
	}
	response.WriteAsJson(result)
}
func logresponse(resp *restful.Response, code int, message string) {
	resp.WriteErrorString(code, message)
	log.Print(message)
}

func main() {
	pwd, _ := os.Getwd()
	log.Println("Starting up in " + pwd)

	ws := new(restful.WebService).
		Produces(restful.MIME_JSON).
		Consumes(restful.MIME_JSON)

	//BEGIN: CORS support
	if enable_cors := os.Getenv("ENABLE_CORS"); enable_cors != "" {
		cors := restful.CrossOriginResourceSharing{
			ExposeHeaders:  []string{"X-My-Header"},
			AllowedHeaders: []string{"Content-Type", "Accept"},
			AllowedMethods: []string{"GET", "POST", "PUT"},
			CookiesAllowed: false,
			Container:      restful.DefaultContainer}

		restful.DefaultContainer.Filter(cors.Filter)
		// Add container filter to respond to OPTIONS
		restful.DefaultContainer.Filter(restful.DefaultContainer.OPTIONSFilter)
	}
	//END: CORS support

	s := &readabilityservice{}
	if r, err := readability.NewReadability("de"); err == nil {
		s.r = r
	} else {
		log.Fatalf("Cannot create NewReadability Instance: %s\n", err.Error())
		return
	}

	ws.Route(ws.PUT("/readability").
		To(s.readabilityservice).
		Produces(restful.MIME_JSON).
		Consumes(restful.MIME_JSON).
		Doc("performs readability checks on an input string").
		Reads(ReadabilityRequest{}).
		Returns(http.StatusOK, "success", ReadabilityResponse{}).
		Returns(http.StatusInternalServerError, "failure", nil).
		Returns(http.StatusBadRequest, "failure", nil))
	ws.Route(ws.PUT("/portalreadability").
		To(s.portalreadabilityservice).
		Produces(restful.MIME_JSON).
		Consumes(restful.MIME_JSON).
		Doc("performs a readability check based on the CKAN AT Open Data Metadata scheme https://www.ref.gv.at/Veroeffentlichte-Informationen.2774.0.html . Only the fields description (ID=9, CKAN \"notes\"), title (ID=8, CKAN \"title\") and keywords (ID=11, CKAN \"Tags\") are used for the check").
		Reads(PortalReadabilityRequest{}).
		Returns(http.StatusOK, "success", PortalReadabilityResponse{}).
		Returns(http.StatusInternalServerError, "failure", nil).
		Returns(http.StatusBadRequest, "failure", nil))
	restful.Add(ws)

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	hostname := os.Getenv("HOSTNAME")

	config := swagger.Config{
		WebServices:     restful.DefaultContainer.RegisteredWebServices(),
		ApiPath:         "/apidocs/apidocs.json",
		SwaggerPath:     "/swagger/",
		SwaggerFilePath: "./swagger-ui/dist"}
	swagger.RegisterSwaggerService(config, restful.DefaultContainer)

	log.Fatal(http.ListenAndServe(hostname+":"+port, nil))
}
