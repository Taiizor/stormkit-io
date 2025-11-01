package subscriptionhandlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/stormkit-io/stormkit-io/src/ce/api/user"
	"github.com/stormkit-io/stormkit-io/src/lib/config"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp"
	"github.com/stormkit-io/stormkit-io/src/lib/slog"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/webhook"
)

var productIDToPackage = map[string]string{
	// test
	"prod_THDhiOfzmRa6xD_test": config.PackagePremium,
	"prod_Rw6no7lokoLIVD_test": config.PackageUltimate,

	// prod
	"prod_THDhiOfzmRa6xD": config.PackagePremium,
	"prod_Rw6no7lokoLIVD": config.PackageUltimate,
}

// handlerSubscriptionUpdate updates the subscription of the user to the given one.
// It also makes sure that a stripe client exists.
func handlerSubscriptionUpdate(req *shttp.RequestContext) *shttp.Response {
	const MaxBodyBytes = int64(65536)
	req.Body = http.MaxBytesReader(req.Writer(), req.Body, MaxBodyBytes)
	payload, err := io.ReadAll(req.Body)

	if err != nil {
		slog.Errorf("error while reading stripe body: %s", err.Error())
		return shttp.Error(err)
	}

	// Pass the request body and Stripe-Signature header to ConstructEvent, along
	// with the webhook signing key.
	event, err := webhook.ConstructEvent(
		payload,
		req.Header.Get("Stripe-Signature"),
		config.Get().Stripe.WebhooksSecret,
	)

	if err != nil {
		slog.Errorf("error while constructing stripe event: %s", err.Error())
		return shttp.Error(err)
	}

	// Testing this locally:
	// $ gem install ultrahook
	// $ export GEM_HOME="$HOME/.gem"
	// $ cd ~/.gem/ruby/2.6.0/bin
	// $ ./ultrahook stripe 8080

	switch event.Type {

	case
		"customer.subscription.created",
		"customer.subscription.updated",
		"customer.subscription.canceled":
		var subscription stripe.Subscription

		if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
			slog.Errorf("error while unmarshaling stripe event: %s", err.Error())
			return shttp.Error(err)
		}

		return UpdateSubscription(req, subscription)

	default:
		return shttp.NoContent()
	}
}

// UpdateSubscription updates the subscription of the user based on the provided subscription info.
func UpdateSubscription(req *shttp.RequestContext, subscription stripe.Subscription) *shttp.Response {
	client := stripeClient()
	customer, err := client.Customers(subscription.Customer.ID, nil)

	if err != nil {
		return shttp.Error(err, fmt.Sprintf("error while retrieving customer: %v", err))
	}

	if customer == nil {
		if subscription.Customer == nil {
			slog.Errorf("error while looking for stripe customer: nil customer in subscription")
		} else {
			slog.Errorf("error while looking for stripe customer: %s", subscription.Customer.ID)
		}

		return shttp.NotFound()
	}

	packageName := config.PackageFree
	quantity := int64(0)

	if subscription.Items != nil {
		for _, sub := range subscription.Items.Data {
			// The last item is usually the new plan, so keep collecting until the end
			// instead of breaking the loop.
			if pck := productIDToPackage[sub.Plan.Product.ID]; pck != "" {
				packageName = pck
				quantity = sub.Quantity
			}
		}
	}

	store := user.NewStore()
	usr, err := store.UserByEmail(req.Context(), []string{customer.Email})

	if err != nil {
		slog.Errorf("error while retrieving user from stripe email: %s, err: %v", customer.Email, err)
		return shttp.Error(err)
	}

	if usr == nil {
		slog.Errorf("user not found with stripe email: %s, customer: %v", customer.Email, customer)
		return shttp.NotFound()
	}

	meta := usr.Metadata
	meta.PackageName = packageName
	meta.StripeCustomerID = customer.ID
	meta.SeatsPurchased = int(quantity)

	if err := store.UpdateSubscription(req.Context(), usr.ID, meta); err != nil {
		slog.Errorf("error while updating subscription: %v", err)
		return shttp.Error(err)
	}

	return shttp.OK()
}
