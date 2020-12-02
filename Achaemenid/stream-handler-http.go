/* For license and copyright information please see LEGAL file in repository */

package achaemenid

import (
	"encoding/base64"
	"strings"

	"../authorization"
	"../convert"
	er "../error"
	"../http"
	"../log"
)

// Indicate standard listen and send port number register for http protocol
const (
	ProtocolPortHTTPReceive uint16 = 80
	ProtocolPortHTTPSend    uint16 = 81
)

// HTTPHandler use to standard HTTP handlers in any layer!
type HTTPHandler func(*Stream, *http.Request, *http.Response)

// HTTPIncomeRequestHandler handle incoming HTTP request streams!
// It can use for architectures like restful, ...
// Protocol Standard - http2 : https://httpwg.org/specs/rfc7540.html
func HTTPIncomeRequestHandler(s *Server, st *Stream) {
	var err *er.Error
	var req = http.MakeNewRequest()
	var res = http.MakeNewResponse()

	err = req.UnMarshal(st.IncomePayload)
	if err != nil {
		if st.Connection == nil {
			// TODO::: due to IPv4&&IPv6 support we must handle some functionality here! Remove it when remove those support.
			// TODO::: attack??
			return
		}
		res.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		res.SetError(err)
		st.Connection.FailedPacketsReceived++
		HTTPOutcomeResponseHandler(s, st, req, res)
		return
	}

	// TODO::: due to IPv4&&IPv6 support we must handle some functionality here! Remove them when remove those support.
	if st.Connection == nil {
		// HTTP may transmit over TCP||UDP and we can't make full detail connection in that protocols!!
		err = checkConnectionInHTTP(s, st, req, res)
		if err != nil {
			return
		}
	}

	if hostCheck(s, st, req, res) {
		return
	}

	// Find related services
	if req.URI.Path == "/apis" {
		if req.Method != http.MethodPOST {
			st.Connection.ServiceCallFail()
			res.SetStatus(http.StatusMethodNotAllowedCode, http.StatusMethodNotAllowedPhrase)
			HTTPOutcomeResponseHandler(s, st, req, res)
			return
		}

		var id uint32
		id, err = convert.Base10StringToUint32(req.URI.Query)
		if err == nil {
			st.Service = s.Services.GetServiceHandlerByID(uint32(id))
		}
		// Add some header for /apis like not index by SE(google, ...), ...
		res.Header.Set("X-Robots-Tag", "noindex")
		// res.Header.Set(http.HeaderKeyCacheControl, "no-store")

		// res.Header.Set(http.HeaderKeyContentEncoding, "gzip")
		// var b bytes.Buffer
		// var gz = gzip.NewWriter(&b)
		// gz.Write(res.Body)
		// gz.Close()
		// res.Body = b.Bytes()
	} else if req.URI.Path == "/objects" {
		var file = s.Assets.Objects.GetFile(req.URI.Query)
		if file == nil {
			st.Connection.ServiceCallFail()
			file = s.Assets.WWW.GetFile("404.html")
			res.SetStatus(http.StatusNotFoundCode, http.StatusNotFoundPhrase)
		} else {
			st.Connection.ServiceCallOK()
			res.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
			res.Header.Set(http.HeaderKeyCacheControl, "max-age=31536000, immutable")
			res.Header.Set(http.HeaderKeyContentType, file.MimeType)
			res.Header.Set(http.HeaderKeyContentEncoding, file.CompressType)
			res.Body = file.CompressData
		}
		HTTPOutcomeResponseHandler(s, st, req, res)
		return
	} else {
		// Route by URL
		st.Service = s.Services.GetServiceHandlerByURI(req.URI.Path)

		// Route by WWW assets
		if st.Service == nil {
			var path = strings.Split(req.URI.Path, "/")
			var lastPath = path[len(path)-1]

			var file = s.Assets.WWW.GetFile(lastPath)
			if file == nil && strings.IndexByte(lastPath, '.') == -1 {
				// TODO::: serve-to-robots
				file = s.Assets.WWWMain
			}

			if file == nil {
				st.Connection.ServiceCallFail()
				res.SetStatus(http.StatusNotFoundCode, http.StatusNotFoundPhrase)
			} else {
				st.Connection.ServiceCallOK()
				res.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
				res.Header.Set(http.HeaderKeyContentType, file.MimeType)
				res.Header.Set(http.HeaderKeyCacheControl, "max-age=31536000, immutable")
				res.Header.Set(http.HeaderKeyContentEncoding, file.CompressType)
				res.Body = file.CompressData
			}

			HTTPOutcomeResponseHandler(s, st, req, res)
			return
		}
	}

	// If project don't have any logic that support data on e.g. HTTP (restful, ...) we reject request with related error.
	if st.Service == nil {
		st.Connection.ServiceCallFail()
		res.SetStatus(http.StatusNotFoundCode, http.StatusNotFoundPhrase)
		res.SetError(http.ErrNotFound)
	} else {
		// Handle request stream
		st.Service.HTTPHandler(st, req, res)
		if st.Err != nil {
			st.Connection.ServiceCallFail()
			res.SetError(st.Err)
		}
	}
	st.Connection.ServiceCallOK()
	HTTPOutcomeResponseHandler(s, st, req, res)
}

// HTTPIncomeResponseHandler use to handle incoming HTTP response streams!
func HTTPIncomeResponseHandler(s *Server, st *Stream) {
	st.SetState(StateReady)
}

// HTTPOutcomeRequestHandler use to handle outcoming HTTP request stream!
// given stream can't be nil, otherwise panic will occur!
// It block caller until get response or error!!
func HTTPOutcomeRequestHandler(s *Server, st *Stream) (err *er.Error) {
	err = st.SendAndWait()
	return
}

// HTTPOutcomeResponseHandler use to handle outcoming HTTP response stream!
func HTTPOutcomeResponseHandler(s *Server, st *Stream, req *http.Request, res *http.Response) (err *er.Error) {
	st.Connection.StreamPool.CloseStream(st)

	// Do some global assignment to response
	res.Version = req.Version
	// res.Header.Set(http.HeaderKeyAccessControlAllowOrigin, "*")
	res.SetContentLength()
	// Add Server Header to response : "Achaemenid"
	res.Header.Set(http.HeaderKeyServer, http.DefaultServer)

	st.OutcomePayload = res.Marshal()

	// send stream to outcome pool by weight
	err = st.Send()
	// TODO::: handle server error almost due to no network available or connection closed!

	return
}

// TODO::: Have default error pages and can get customizes!
// Send beauty HTML response in http error situation like 500, 404, ...

// HTTPToHTTPSHandler handle incoming HTTP request streams to redirect to HTTPS!
// TODO::: remove this when remove support of IP!
func HTTPToHTTPSHandler(s *Server, st *Stream) {
	var err *er.Error
	var req = http.MakeNewRequest()
	var res = http.MakeNewResponse()
	err = req.UnMarshal(st.IncomePayload)
	if err != nil {
		st.Connection.ServiceCallFail()
		res.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		// res.SetError(err)
	} else {
		// redirect http to https
		// remove/add not default ports from req.Host
		var target = "https://" + req.GetHost() + req.URI.Path
		if len(req.URI.Query) > 0 {
			target += "?" + req.URI.Query // + "&rd=tls" // TODO::: add rd query for analysis purpose??
		}
		res.SetStatus(http.StatusMovedPermanentlyCode, http.StatusMovedPermanentlyPhrase)
		res.Header.Set(http.HeaderKeyLocation, target)
		res.Header.Set(http.HeaderKeyConnection, http.HeaderValueClose)
		// Add cache to decrease server load
		res.Header.Set(http.HeaderKeyCacheControl, "public, max-age=2592000")
	}

	st.Connection.ServiceCallOK()

	// Do some global assignment to response
	res.Version = req.Version
	res.Header.Set(http.HeaderKeyContentLength, "0")
	// Add Server Header to response : "Achaemenid"
	res.Header.Set(http.HeaderKeyServer, http.DefaultServer)

	st.OutcomePayload = res.Marshal()
}

// HTTP Cookie some names and
// TODO::: due to IPv4&&IPv6 support need these const! Remove them when remove those support.
const (
	HTTPCookieNameBaseUserID = "B"  // any of users type ID indicate in HTTPConnAllowBaseUsers
	HTTPCookieNameBaseConnID = "BC" // Achaemenid Base Connection ID

	HTTPCookieNameDelegateUserID = "D"  // Any User Type ID
	HTTPCookieNameDelegateConnID = "DC" // Achaemenid Delegate Connection ID

	HTTPCookieNameThingID = "T" // Achaemenid Thing ID

	HTTPCookieValueGuestUserID = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA" // Guest UserID == [32]byte{}

	HTTPConnAllowBaseUsers = authorization.UserTypePerson // | authorization.UserTypeAI | authorization.UserTypeApp
)

// TODO::: due to IPv4&&IPv6 support need this func! Remove them when remove those support.
func checkConnectionInHTTP(s *Server, st *Stream, req *http.Request, res *http.Response) (err *er.Error) {
	var goErr error
	// First try to find cookie on http header!
	var cookies = req.Header.GetCookies()
	var personID, personConnID, delUserID, delConnID, ThingID [32]byte
	if len(cookies) > 1 {
		var ln = len(cookies) // TODO::: Is it overkilling code to get len seprately???
		for i := 0; i < ln; i++ {
			switch cookies[i].Name {
			case HTTPCookieNameBaseUserID:
				_, goErr = base64.RawStdEncoding.Decode(personID[:], convert.UnsafeStringToByteSlice(cookies[i].Value))
			case HTTPCookieNameBaseConnID:
				_, goErr = base64.RawStdEncoding.Decode(personConnID[:], convert.UnsafeStringToByteSlice(cookies[i].Value))

			case HTTPCookieNameDelegateUserID:
				_, goErr = base64.RawStdEncoding.Decode(delUserID[:], convert.UnsafeStringToByteSlice(cookies[i].Value))
			case HTTPCookieNameDelegateConnID:
				_, goErr = base64.RawStdEncoding.Decode(delConnID[:], convert.UnsafeStringToByteSlice(cookies[i].Value))

			case HTTPCookieNameThingID:
				_, goErr = base64.RawStdEncoding.Decode(ThingID[:], convert.UnsafeStringToByteSlice(cookies[i].Value))
			}
			// Don't stop loop even one of decoding occur error!
		}
		if goErr == nil && personID != [32]byte{} && personConnID != [32]byte{} {
			// Anyway first check given person details is valid
			st.Connection = s.Connections.GetConnectionByID(personConnID)
			if st.Connection != nil && (st.Connection.UserID != personID || st.Connection.UserType.CheckReverse(HTTPConnAllowBaseUsers) != nil) {
				// Attack!!
				// err =
				return
			}

			if st.Connection != nil {
				// If any org details exist change connection to org one!
				if delUserID != [32]byte{} && delConnID != [32]byte{} && delUserID != personID {
					var delegateConn *Connection
					delegateConn = s.Connections.GetConnectionByID(delConnID)
					if delegateConn != nil && (delegateConn.UserID != delUserID || delegateConn.DelegateUserID != personID) {
						// Attack!!
						// err =
						return
					}
					if delegateConn != nil {
						st.Connection = delegateConn
					}
				}
			}
		}
	}

	// If can't find or get exiting connection from local and remote, make new one if server allow!
	if st.Connection == nil {
		st.Connection, err = s.Connections.MakeNewGuestConnection()
		if err != nil {
			// Can't make new guest connection almost due to lack of enough resources, So simply return to close the connection
			return
		}

		var cookies = []http.SetCookie{
			http.SetCookie{
				Name:     HTTPCookieNameBaseConnID,
				Value:    base64.RawStdEncoding.EncodeToString(st.Connection.ID[:]),
				MaxAge:   "630720000", // = 20 year = 20*365*24*60*60
				Secure:   true,
				HTTPOnly: true,
				SameSite: "Lax",
			}, http.SetCookie{
				Name:     HTTPCookieNameBaseUserID,
				Value:    HTTPCookieValueGuestUserID,
				MaxAge:   "630720000", // = 20 year = 20*365*24*60*60
				Secure:   true,
				HTTPOnly: false,
				SameSite: "Lax",
			},
		}
		if log.DevMode {
			cookies[0].Secure = false
			cookies[1].Secure = false
		}
		res.Header.SetSetCookies(cookies)

		// It is first time user reach platform, So tell the browser always reach by https!
		res.Header.Set(http.HeaderKeyStrictTransportSecurity, "max-age=31536000; includeSubDomains; preload")
	}

	st.Connection.SetThingID(ThingID)
	s.Connections.RegisterConnection(st.Connection)

	// TODO::: This header just need in IP connections, so add here not globally in HTTPOutcomeResponseHandler()
	res.Header.Set(http.HeaderKeyConnection, http.HeaderValueKeepAlive)
	res.Header.Set(http.HeaderKeyKeepAlive, "timeout="+tcpKeepAliveDurationString)

	st.Connection.StreamPool.RegisterStream(st)
	return
}

// TODO::: due to IPv4&&IPv6 support need this func! Remove them when remove those support.
func hostCheck(s *Server, st *Stream, req *http.Request, res *http.Response) (redirect bool) {
	if !log.DevMode {
		var host = req.GetHost()

		if host == "" {
			// TODO::: noting to do or reject request??
		} else if '0' <= host[0] && host[0] <= '9' {
			// check of request send over IP
			if log.DebugMode {
				log.Debug("HTTP - Host Check - IP host:", host)
			}

			// Add www to domain. Just support http on www server app due to SE duplicate content both on www && non-www!
			var target = "https://" + s.Manifest.DomainName + req.URI.Path
			if len(req.URI.Query) > 0 {
				target += "?" + req.URI.Query // + "&rd=tls" // TODO::: add rd query for analysis purpose??
			}
			res.SetStatus(http.StatusMovedPermanentlyCode, http.StatusMovedPermanentlyPhrase)
			res.Header.Set(http.HeaderKeyLocation, target)
			res.Header.Set(http.HeaderKeyCacheControl, "max-age=31536000, immutable")
			HTTPOutcomeResponseHandler(s, st, req, res)
			return true
		} else if len(host) > 4 && host[:4] == "www." {
			if host[4:] != s.Manifest.DomainName {
				if log.DebugMode {
					log.Debug("HTTP - Host Check - Unknown WWW host:", host)
				}
				// TODO::: Silently ignoring a request might not be a good idea and perhaps breaks the RFC's for HTTP.
				return true
			}

			if log.DebugMode {
				log.Debug("HTTP - Host Check - WWW host:", host)
			}

			// Add www to domain. Just support http on www server app due to SE duplicate content both on www && non-www!
			var target = "https://" + s.Manifest.DomainName + req.URI.Path
			if len(req.URI.Query) > 0 {
				target += "?" + req.URI.Query // + "&rd=tls" // TODO::: add rd query for analysis purpose??
			}
			res.SetStatus(http.StatusMovedPermanentlyCode, http.StatusMovedPermanentlyPhrase)
			res.Header.Set(http.HeaderKeyLocation, target)
			res.Header.Set(http.HeaderKeyCacheControl, "max-age=31536000, immutable")
			HTTPOutcomeResponseHandler(s, st, req, res)
			return true
		} else if host != s.Manifest.DomainName {
			if log.DebugMode {
				log.Debug("HTTP - Host Check - Unknown host:", host)
			}
			// TODO::: Silently ignoring a request might not be a good idea and perhaps breaks the RFC's for HTTP.
			return true
		}
	}
	return
}
