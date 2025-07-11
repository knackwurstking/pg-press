package pgvis

type TroubleReport struct {
	Modified *Modified[*TroubleReport] `json:"modified"`
	Title    string                    `json:"title"`
	Content  string                    `json:"content"`
}

func NewTroubleReport(m *Modified[*TroubleReport], title, content string) *TroubleReport {
	return &TroubleReport{
		Modified: m,
	}
}

// TODO: Create DBTroubleReports
