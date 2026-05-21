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
	// I2568C001: BG accepted v1; PO now at v2 with changed sub-job numbers → has_update
	{ProjectCode: "I2568C001", Version: 2, UpdatedAt: time.Date(2026, 5, 19, 10, 0, 0, 0, time.UTC)},
	// I2570C003: BG has budget_sources only, PO has sub_jobs → new
	{ProjectCode: "I2570C003", Version: 1, UpdatedAt: time.Date(2026, 5, 18, 8, 0, 0, 0, time.UTC)},
	// I2570C004: BG has budget_sources only, PO also has no sub_jobs → up_to_date
	{ProjectCode: "I2570C004", Version: 1, UpdatedAt: time.Date(2026, 5, 15, 8, 0, 0, 0, time.UTC)},
	// I2570C005: no BG record at all, starts at active year → new
	{ProjectCode: "I2570C005", Version: 1, UpdatedAt: time.Date(2026, 5, 20, 9, 0, 0, 0, time.UTC)},
}

var str = func(s string) *string { return &s }

var mockProjects = map[string]models.POProject{
	// I2568C001: same sub_job names as BG; หมวดสิ่งก่อสร้าง and หมวดยานพาหนะ ลงทุน have changed
	"I2568C001": {
		ProjectCode: "I2568C001",
		Version:     2,
		ExportedAt:  time.Date(2026, 5, 19, 10, 0, 0, 0, time.UTC),
		Name:        "แผนงานสนับสนุนการดำเนินงาน ระยะที่ 6",
		Division:    str("รผก.(วว-วศ)/ฝยธ."),
		Department:  str("วว"),
		SubJobs: []models.POSubJob{
			{Name: "1. หมวดที่ดิน",               FundType: "ลงทุน",  DataYear: 2570, Budget: 0, Target: 0},
			{Name: "2. หมวดสิ่งก่อสร้าง",         FundType: "ผูกพัน", DataYear: 2570, Budget: 1324.969, Target: 173.955},
			{Name: "2. หมวดสิ่งก่อสร้าง",         FundType: "ลงทุน",  DataYear: 2570, Budget: 520.000, Target: 34.114},
			{Name: "3. หมวดยานพาหนะ",             FundType: "ผูกพัน", DataYear: 2570, Budget: 820.922, Target: 468.5},
			{Name: "3. หมวดยานพาหนะ",             FundType: "ลงทุน",  DataYear: 2570, Budget: 850.000, Target: 180.91},
			{Name: "4. หมวดเครื่องมือ-เครื่องใช้", FundType: "ผูกพัน", DataYear: 2570, Budget: 620.592, Target: 255.765},
			{Name: "4. หมวดเครื่องมือ-เครื่องใช้", FundType: "ลงทุน",  DataYear: 2570, Budget: 345.04, Target: 147.07},
			{Name: "สำรองราคา",                   FundType: "ผูกพัน", DataYear: 2570, Budget: 152.117, Target: 0},
		},
	},
	// I2570C003: BG has budget_sources only; PO has sub_jobs → all will appear as "added"
	"I2570C003": {
		ProjectCode: "I2570C003",
		Version:     1,
		ExportedAt:  time.Date(2026, 5, 18, 8, 0, 0, 0, time.UTC),
		Name:        "แผนงานการก่อสร้างปรับปรุงระบบจำหน่ายไฟฟ้าเป็นเคเบิลใต้ดิน",
		Division:    str("ฝบก.1"),
		Department:  str("ป"),
		SubJobs: []models.POSubJob{
			{Name: "งานรวม", FundType: "ลงทุน", DataYear: 2570, Budget: 850.000, Target: 78.45},
			{Name: "งานรวม", FundType: "ลงทุน", DataYear: 2571, Budget: 920.000, Target: 0},
		},
	},
	// I2570C004: BG has budget_sources only; PO also has no sub_jobs → diff is empty, up_to_date
	"I2570C004": {
		ProjectCode: "I2570C004",
		Version:     1,
		ExportedAt:  time.Date(2026, 5, 15, 8, 0, 0, 0, time.UTC),
		Name:        "แผนงานเช่ายานพาหนะ ปี 2570 - 2575",
		Division:    str("รผก.(ล)"),
		Department:  str("ล"),
		SubJobs:     []models.POSubJob{},
	},
	// I2570C005: no BG record at all, new project starting at active year
	"I2570C005": {
		ProjectCode: "I2570C005",
		Version:     1,
		ExportedAt:  time.Date(2026, 5, 20, 9, 0, 0, 0, time.UTC),
		Name:        "แผนงานพัฒนาระบบพลังงานทดแทน ระยะที่ 1",
		Division:    str("ฝวพ."),
		Department:  str("พน"),
		SubJobs: []models.POSubJob{
			{Name: "1. งานสำรวจและออกแบบ", FundType: "ลงทุน", DataYear: 2570, Budget: 15000.000, Target: 0},
			{Name: "2. งานก่อสร้าง",       FundType: "ลงทุน", DataYear: 2570, Budget: 85000.000, Target: 0},
			{Name: "2. งานก่อสร้าง",       FundType: "ลงทุน", DataYear: 2571, Budget: 120000.000, Target: 0},
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
