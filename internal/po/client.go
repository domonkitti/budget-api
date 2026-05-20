package po

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/domonkitti/budget-app-api/internal/models"
)

// Client is the interface for fetching data from the PO system.
type Client interface {
	FetchVersions(ctx context.Context) ([]models.POProjectVersion, error)
	FetchProject(ctx context.Context, projectCode string) (*models.POProject, error)
}

// --- HTTP implementation ---

type httpClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewHTTPClient(baseURL string) Client {
	return &httpClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *httpClient) FetchVersions(ctx context.Context) ([]models.POProjectVersion, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/export/projects/versions", nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("po api versions: status %d", resp.StatusCode)
	}
	var result struct {
		Projects []models.POProjectVersion `json:"projects"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Projects, nil
}

func (c *httpClient) FetchProject(ctx context.Context, projectCode string) (*models.POProject, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/export/project/"+projectCode, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("project %s not found in PO system", projectCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("po api project: status %d", resp.StatusCode)
	}
	var p models.POProject
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		return nil, err
	}
	return &p, nil
}

// --- Mock implementation ---

type mockClient struct{}

func NewMockClient() Client {
	return &mockClient{}
}

var mockVersions = []models.POProjectVersion{
	{ProjectCode: "I2570Y001", Version: 3, UpdatedAt: time.Date(2026, 5, 18, 9, 0, 0, 0, time.UTC)},
	{ProjectCode: "I2570Y002", Version: 2, UpdatedAt: time.Date(2026, 4, 10, 14, 0, 0, 0, time.UTC)},
	{ProjectCode: "I2570C001", Version: 1, UpdatedAt: time.Date(2026, 5, 15, 11, 0, 0, 0, time.UTC)},
}

var str = func(s string) *string { return &s }

var mockProjects = map[string]models.POProject{
	"I2570Y001": {
		ProjectCode: "I2570Y001",
		Version:     3,
		ExportedAt:  time.Date(2026, 5, 18, 9, 0, 0, 0, time.UTC),
		Name:        "โครงการปรับปรุงระบบน้ำประปา (ฉบับปรับปรุง)",
		Division:    str("สำนักสาธารณูปโภค"),
		Department:  str("ฝ่ายโครงสร้างพื้นฐาน"),
		GroupName:   str("หมวดสาธารณูปโภค"),
		SubJobs: []models.POSubJob{
			{Name: "งานออกแบบ", FundType: "ลงทุน", DataYear: 2570, Budget: 1500000, Target: 1200000},
			{Name: "งานก่อสร้าง", FundType: "ลงทุน", DataYear: 2570, Budget: 8000000, Target: 7500000},
			{Name: "งานก่อสร้าง", FundType: "ผูกพัน", DataYear: 2570, Budget: 2000000, Target: 1800000},
			{Name: "งานตรวจสอบ", FundType: "ลงทุน", DataYear: 2570, Budget: 500000, Target: 400000},
		},
	},
	"I2570Y002": {
		ProjectCode: "I2570Y002",
		Version:     2,
		ExportedAt:  time.Date(2026, 4, 10, 14, 0, 0, 0, time.UTC),
		Name:        "โครงการก่อสร้างอาคารสำนักงาน",
		Division:    str("สำนักโยธา"),
		Department:  str("ฝ่ายก่อสร้าง"),
		GroupName:   str("หมวดก่อสร้าง"),
		SubJobs: []models.POSubJob{
			{Name: "งานฐานราก", FundType: "ลงทุน", DataYear: 2570, Budget: 3000000, Target: 2800000},
			{Name: "งานโครงสร้าง", FundType: "ลงทุน", DataYear: 2570, Budget: 12000000, Target: 11000000},
		},
	},
	"I2570C001": {
		ProjectCode: "I2570C001",
		Version:     1,
		ExportedAt:  time.Date(2026, 5, 15, 11, 0, 0, 0, time.UTC),
		Name:        "โครงการพัฒนาระบบสารสนเทศ",
		Division:    str("สำนักเทคโนโลยี"),
		Department:  str("ฝ่ายพัฒนาระบบ"),
		SubJobs: []models.POSubJob{
			{Name: "งานวิเคราะห์ความต้องการ", FundType: "ลงทุน", DataYear: 2570, Budget: 800000, Target: 700000},
			{Name: "งานพัฒนาระบบ", FundType: "ลงทุน", DataYear: 2570, Budget: 4500000, Target: 4000000},
		},
	},
}

func (m *mockClient) FetchVersions(_ context.Context) ([]models.POProjectVersion, error) {
	return mockVersions, nil
}

func (m *mockClient) FetchProject(_ context.Context, projectCode string) (*models.POProject, error) {
	p, ok := mockProjects[projectCode]
	if !ok {
		return nil, fmt.Errorf("project %s not found in PO system", projectCode)
	}
	return &p, nil
}
