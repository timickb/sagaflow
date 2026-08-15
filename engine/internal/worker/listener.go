package worker

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

const notifyChannel = "saga_runnable"

// Listener держит отдельное соединение и слушает NOTIFY,
// превращая сигналы в неблокирующие "пинки" диспетчеру
type Listener struct {
	dsn  string
	wake chan struct{}
}

func NewListener(dsn string) *Listener {
	return &Listener{
		dsn: dsn,
		// буфер 1 = дедупликация: пока диспетчер не разгреб прошлый
		// сигнал, лавина NOTIFY схлопывается в один (anti thundering herd).
		wake: make(chan struct{}, 1),
	}
}

// Wake — канал, из которого диспетчер читает "есть работа"
func (l *Listener) Wake() <-chan struct{} {
	return l.wake
}

// Run слушает до отмены контекста, переподключаясь при обрывах соединения
func (l *Listener) Run(ctx context.Context) {
	for {
		if err := l.listen(ctx); err != nil {
			if ctx.Err() != nil {
				return // штатное завершение
			}
			log.Error().Err(err).Msg("Listener lost connection, reconnecting")
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second): // backoff
			}
		}
	}
}

func (l *Listener) listen(ctx context.Context) error {
	conn, err := pgx.Connect(ctx, l.dsn)
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())

	if _, err := conn.Exec(ctx, "LISTEN "+notifyChannel); err != nil {
		return err
	}
	log.Info().Msg("Listening for saga_runnable notifications")

	for {
		// блокируется на сокете до нотификации либо отмены ctx — без поллинга
		if _, err := conn.WaitForNotification(ctx); err != nil {
			return err
		}
		l.signal()
	}
}

// неблокирующая отправка: если сигнал уже висит непрочитанным — отбрасываем
func (l *Listener) signal() {
	select {
	case l.wake <- struct{}{}:
	default:
	}
}
