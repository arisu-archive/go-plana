package plana_test

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/arisu-archive/plana-protos/protos"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/arisu-archive/go-plana/plana"
)

var _ = Describe("Request cancellation", func() {
	It("propagates the request context through Client.Do", func() {
		ctx, cancel := context.WithCancel(context.Background())
		DeferCleanup(cancel)

		transportErr := errors.New("request context was not canceled")
		client := plana.NewClient(nil, &http.Client{
			Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
				cancel()
				select {
				case <-req.Context().Done():
					return nil, req.Context().Err()
				default:
					return nil, transportErr
				}
			}),
		})

		client.GatewayURL = &url.URL{Scheme: "https", Host: "example.invalid"}
		request, err := client.R().WithGatewayBypass().Gateway(
			ctx,
			protos.Protocol_Queuing_GetTicket,
			plana.QueuingGetAuthTicketRequestWrapper{
				QueuingGetAuthTicketRequest: &protos.QueuingGetAuthTicketRequest{},
			},
		)
		Expect(err).NotTo(HaveOccurred())

		_, err = client.Do(request, nil)
		Expect(errors.Is(err, context.Canceled)).To(BeTrue(), "Client.Do() error = %v", err)
	})
})

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
