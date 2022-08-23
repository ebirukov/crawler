package crawler

type MetricMock struct{}

func (m MetricMock) IncDuplicate() {}

func (m MetricMock) IncProcessed() {}

func (m MetricMock) IncSkipped(_ int) {}

func (m MetricMock) IncSubmitted() {}

func (m MetricMock) IncRequestTimeout() {}
