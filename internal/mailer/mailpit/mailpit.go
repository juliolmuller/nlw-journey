package mailpit

import (
	"context"
	"fmt"
	"nlw-journey/internal/pgstore"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/wneessen/go-mail"
)

type Store interface {
	GetTrip(context.Context, uuid.UUID) (pgstore.Trip, error)
}

type Mailpit struct {
	store Store
}

func NewMailpit(pool *pgxpool.Pool) Mailpit {
	store := pgstore.New(pool)

	return Mailpit{store}
}

func (mp Mailpit) SendConfirmTripEmailToTripOwner(tripID uuid.UUID) error {
	ctx := context.Background()
	trip, err := mp.store.GetTrip(ctx, tripID)
	if err != nil {
		return fmt.Errorf("mailpit: fail to get trip in SendConfirmTripEmailToTripOwner; %s", err)
	}

	msg := mail.NewMsg()
	if err := msg.From("no-replay@nlw-journey.com"); err != nil {
		return fmt.Errorf("mailpit: fail to set 'from' property in SendConfirmTripEmailToTripOwner; %s", err)
	}
	if err := msg.To(trip.OwnerEmail); err != nil {
		return fmt.Errorf("mailpit: fail to set 'to' property in SendConfirmTripEmailToTripOwner; %s", err)
	}
	msg.Subject("Confirm Your Trip")
	msg.SetBodyString(mail.TypeTextHTML, fmt.Sprintf(`
	<h2>Hello, %s,</h2>
	<p>Your trip to %s on %s must be confirmed.</p>
	<a href="#">Confirm Trip</a>
	`, trip.OwnerName, trip.Destination, trip.StartsAt.Time.Format(time.DateOnly)))

	client, err := mail.NewClient("localhost", mail.WithTLSPortPolicy(mail.NoTLS), mail.WithPort(1025))
	if err != nil {
		return fmt.Errorf("mailpit: fail to create mail client in SendConfirmTripEmailToTripOwner; %s", err)
	}

	if err := client.DialAndSend(msg); err != nil {
		return fmt.Errorf("mailpit: fail to send email in SendConfirmTripEmailToTripOwner; %s", err)
	}

	return nil
}
