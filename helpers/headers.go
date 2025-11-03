package helpers

import "net/http"

// CopyHeaders copies values from the source to the destination headers map.
func CopyHeaders(dest http.Header, src http.Header) {
	// NOTE: yaegi doesn't properly support generics for copy
	//		 so cannot use maps.Copy
	for k, v := range src {
		dest[k] = v
	}
}
