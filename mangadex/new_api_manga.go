package mangadex

import (
	"context"
	"github.com/antihax/optional"
	"io"
	"net/http"
	"net/url"
	"strings"
)

var (
	_ context.Context
)

type MangaApiService2 service

type MangaApiGetSearchMangaOpts2 struct {
	Limit          optional.Int32
	Offset         optional.Int32
	Ids            optional.Interface
	OrderCreatedAt optional.String
	UpdatedAtSince optional.String
}

func (a *MangaApiService) GetSearchManga2(ctx context.Context, localVarOptionals *MangaApiGetSearchMangaOpts2) (MangaList, *http.Response, error) {
	var (
		localVarHttpMethod  = strings.ToUpper("Get")
		localVarPostBody    interface{}
		localVarFileName    string
		localVarFileBytes   []byte
		localVarReturnValue MangaList
	)

	// create path and map variables
	localVarPath := a.client.cfg.BasePath + "/manga"

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	localVarQueryParams.Add("contentRating[]", parameterToString("safe", ""))
	localVarQueryParams.Add("contentRating[]", parameterToString("suggestive", ""))
	localVarQueryParams.Add("contentRating[]", parameterToString("erotica", ""))
	localVarQueryParams.Add("contentRating[]", parameterToString("pornographic", ""))

	if localVarOptionals != nil && localVarOptionals.OrderCreatedAt.IsSet() {
		localVarQueryParams.Add("order[createdAt]", parameterToString(localVarOptionals.OrderCreatedAt.Value(), ""))
	}

	if localVarOptionals != nil && localVarOptionals.UpdatedAtSince.IsSet() {
		localVarQueryParams.Add("updatedAtSince", parameterToString(localVarOptionals.UpdatedAtSince.Value(), ""))
	}

	if localVarOptionals != nil && localVarOptionals.Limit.IsSet() {
		localVarQueryParams.Add("limit", parameterToString(localVarOptionals.Limit.Value(), ""))
	}
	if localVarOptionals != nil && localVarOptionals.Offset.IsSet() {
		localVarQueryParams.Add("offset", parameterToString(localVarOptionals.Offset.Value(), ""))
	}

	if localVarOptionals != nil && localVarOptionals.Ids.IsSet() {
		ids := localVarOptionals.Ids.Value().([]string)
		for _, id := range ids {
			localVarQueryParams.Add("ids[]", parameterToString(id, ""))
		}
		//localVarQueryParams.Add("ids[]", parameterToString(localVarOptionals.Ids.Value(), "multi"))
	}

	// to determine the Content-Type header
	localVarHttpContentTypes := []string{}

	// set Content-Type header
	localVarHttpContentType := selectHeaderContentType(localVarHttpContentTypes)
	if localVarHttpContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHttpContentType
	}

	// to determine the Accept header
	localVarHttpHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHttpHeaderAccept := selectHeaderAccept(localVarHttpHeaderAccepts)
	if localVarHttpHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHttpHeaderAccept
	}
	r, err := a.client.prepareRequest(ctx, localVarPath, localVarHttpMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, localVarFileName, localVarFileBytes)

	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHttpResponse, err := a.client.callAPI(r)
	if err != nil || localVarHttpResponse == nil {
		return localVarReturnValue, localVarHttpResponse, err
	}

	localVarBody, err := io.ReadAll(localVarHttpResponse.Body)
	localVarHttpResponse.Body.Close()
	if err != nil {
		return localVarReturnValue, localVarHttpResponse, err
	}
	//fmt.Println(localVarHttpResponse.Header)
	//fmt.Println("X-RateLimit-Limit: " + localVarHttpResponse.Header.Get("X-RateLimit-Limit"))
	//fmt.Println("X-RateLimit-Remaining: " + localVarHttpResponse.Header.Get("X-RateLimit-Remaining"))
	//fmt.Println("X-RateLimit-Retry-After: " + localVarHttpResponse.Header.Get("X-RateLimit-Retry-After"))

	if localVarHttpResponse.StatusCode < 300 {
		// If we succeed, return the data, otherwise pass on to decode error.
		err = a.client.decode(&localVarReturnValue, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
		if err == nil {
			return localVarReturnValue, localVarHttpResponse, err
		}
	}

	if localVarHttpResponse.StatusCode >= 300 {
		newErr := GenericSwaggerError{
			body:  localVarBody,
			error: localVarHttpResponse.Status,
		}
		if localVarHttpResponse.StatusCode == 200 {
			var v MangaList
			err = a.client.decode(&v, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHttpResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHttpResponse, newErr
		}
		if localVarHttpResponse.StatusCode == 400 {
			var v ErrorResponse
			err = a.client.decode(&v, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHttpResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHttpResponse, newErr
		}
		return localVarReturnValue, localVarHttpResponse, newErr
	}

	return localVarReturnValue, localVarHttpResponse, nil
}
