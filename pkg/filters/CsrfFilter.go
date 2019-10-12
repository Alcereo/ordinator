package filters

import (
	"fmt"
	"github.com/Alcereo/ordinator/pkg/common"
	"github.com/Alcereo/ordinator/pkg/crypt"
	"github.com/sirupsen/logrus"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
)

type methodsSet struct {
	methodsMap map[string]bool
}

func newMethodsSet(methods []string) *methodsSet {
	methodsMap := make(map[string]bool)
	for _, method := range methods {
		methodsMap[method] = true
	}
	return &methodsSet{methodsMap: methodsMap}
}

func (set *methodsSet) Contains(method string) bool {
	return set.methodsMap[method] == true
}

type CsrfFilter struct {
	next           *common.RequestHandler
	Name           string           `validate:"required"`
	HeaderName     string           `validate:"required"`
	SafeMethodsSet *methodsSet      `validate:"required"`
	Encryptor      *crypt.Encryptor `validate:"required"`
}

var validate = validator.New()

func NewCsrfFilter(name string, headerName string, safeMethods []string, encryptorPrivateKey string) *CsrfFilter {
	filter := &CsrfFilter{
		Name:           name,
		HeaderName:     headerName,
		SafeMethodsSet: newMethodsSet(safeMethods),
		Encryptor:      crypt.NewEncryptor(encryptorPrivateKey),
	}
	err := validate.Struct(filter)
	if err != nil {
		panic(err.Error())
	}
	return filter
}

func (filter *CsrfFilter) SetNext(nextHandler common.RequestHandler) {
	filter.next = &nextHandler
}

func (filter *CsrfFilter) Handle(log *logrus.Entry, writer http.ResponseWriter, request *http.Request) {
	const stage = "Csrf filter error. Reason: %v"
	log = log.WithField("filterName", filter.Name)

	session, err := resolveSession(request)
	if err != nil {
		log.Errorf(stage, err)
		writer.WriteHeader(500)
		_, _ = fmt.Fprintf(writer, stage, err.Error())
		return
	}

	if !filter.methodIsSafe(request.Method) {
		csrfHeader, err := filter.resolveCsrfHeader(request.Header)
		if err != nil {
			writer.WriteHeader(403)
			_, _ = fmt.Fprintf(writer, err.Error())
			return
		}
		if err := filter.checkCsrfHeader(csrfHeader, session); err != nil {
			writer.WriteHeader(403)
			_, _ = fmt.Fprintf(writer, err.Error())
			return
		}
	} else {
		newCsrfToken, err := filter.generateNewCsrfToken(session)
		if err != nil {
			log.Errorf(stage, err.Error())
			writer.WriteHeader(500)
			_, _ = fmt.Fprintf(writer, stage, err.Error())
			return
		}
		writer.Header().Add(filter.HeaderName, newCsrfToken)
	}
	(*filter.next).Handle(log, writer, request)
}

func (filter *CsrfFilter) methodIsSafe(method string) bool {
	return filter.SafeMethodsSet.Contains(method)
}

func (filter *CsrfFilter) checkCsrfHeader(csrfHeader string, session *common.Session) error {
	value, err := filter.Encryptor.DecryptFact(csrfHeader)
	if err != nil {
		return fmt.Errorf("decrypt CSRF header error. Reason: %v", err.Error())
	}
	if value != string(session.Id) {
		return fmt.Errorf("invalid CSRF token")
	}
	return nil
}

func (filter *CsrfFilter) generateNewCsrfToken(session *common.Session) (string, error) {
	token, err := filter.Encryptor.EncryptFact(string(session.Id))
	if err != nil {
		return "", fmt.Errorf("generation new CSRF token error. Reason: %v", err.Error())
	}
	return token, nil
}

func (filter *CsrfFilter) resolveCsrfHeader(headers http.Header) (string, error) {
	header := headers.Get(filter.HeaderName)
	if header == "" {
		return "", fmt.Errorf("resolving CSRF header error. CSRF header: %v is empty", filter.HeaderName)
	}
	return header, nil
}

func resolveSession(request *http.Request) (*common.Session, error) {
	sessionNillable := request.Context().Value(common.SessionContextKey)
	if sessionNillable == nil {
		return nil, fmt.Errorf("session not found. Session filter required to be performed before CSRF gilter")
	}
	return sessionNillable.(*common.Session), nil
}
