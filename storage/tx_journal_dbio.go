package storage

type TXJournalDBIO struct {
	db Storage
}

func NewTXJournalDBIO(db Storage) *TXJournalDBIO {
	return &TXJournalDBIO{db: db}
}
