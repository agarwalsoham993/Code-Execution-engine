package worker

import (
	"code-runner/internal/config"
	"code-runner/internal/database"
	"code-runner/internal/queue"
	"code-runner/internal/sandbox"
	"github.com/zekrotja/rogu/log"
	"sync"
	"time"
)

type Pool struct {
	cfg     *config.EnvProvider
	queue   *queue.RedisQueue
	db      *database.PostgresDB
	mgr     *sandbox.Manager
	
	workers map[int]*Worker
	mu      sync.Mutex
	nextID  int
}

func NewPool(cfg *config.EnvProvider, q *queue.RedisQueue, db *database.PostgresDB, mgr *sandbox.Manager) *Pool {
	return &Pool{
		cfg:     cfg,
		queue:   q,
		db:      db,
		mgr:     mgr,
		workers: make(map[int]*Worker),
		nextID:  1,
	}
}

func (p *Pool) Start() {
	min := p.cfg.Config().Worker.Min
	log.Info().Msgf("Starting Worker Pool (Min: %d, Max: %d)", min, p.cfg.Config().Worker.Max)

	for i := 0; i < min; i++ {
		p.addWorker()
	}

	go p.autoscaler()
}

func (p *Pool) autoscaler() {
	ticker := time.NewTicker(3 * time.Second)
	for range ticker.C {
		qLen, err := p.queue.Length()
		if err != nil {
			log.Error().Err(err).Msg("Failed to check queue length")
			continue
		}

		p.mu.Lock()
		currentCount := len(p.workers)
		p.mu.Unlock()

		// Simple Algo: 1 worker per 2 pending jobs + buffer, clamped by Min/Max
		desired := p.cfg.Config().Worker.Min + int(qLen/2)
		if desired > p.cfg.Config().Worker.Max {
			desired = p.cfg.Config().Worker.Max
		}

		// 3. Scale Up or Down
		if desired > currentCount {
			log.Info().Msgf("Scaling UP: Queue=%d, Current=%d, Desired=%d", qLen, currentCount, desired)
			toAdd := desired - currentCount
			for i := 0; i < toAdd; i++ {
				p.addWorker()
			}
		} else if desired < currentCount && currentCount > p.cfg.Config().Worker.Min {
			log.Info().Msgf("Scaling DOWN: Queue=%d, Current=%d, Desired=%d", qLen, currentCount, desired)
			toRemove := currentCount - desired
			if currentCount-toRemove < p.cfg.Config().Worker.Min {
				toRemove = currentCount - p.cfg.Config().Worker.Min
			}
			for i := 0; i < toRemove; i++ {
				p.removeWorker()
			}
		}
	}
}

func (p *Pool) addWorker() {
	p.mu.Lock()
	defer p.mu.Unlock()

	id := p.nextID
	p.nextID++
	
	w := NewWorker(id, p.queue, p.db, p.mgr)
	p.workers[id] = w
	go w.Start()
}

func (p *Pool) removeWorker() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for id, w := range p.workers {
		w.Stop()
		delete(p.workers, id)
		return
	}
}