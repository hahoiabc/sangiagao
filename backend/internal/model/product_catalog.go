package model

// RiceCategory represents a category of rice products.
type RiceCategory struct {
	Key      string        `json:"key"`
	Label    string        `json:"label"`
	Products []RiceProduct `json:"products"`
}

// RiceProduct represents a specific rice product within a category.
type RiceProduct struct {
	Key      string `json:"key"`
	Label    string `json:"label"`
	Category string `json:"category"`
}

// RiceCategories is the predefined product catalog.
var RiceCategories = []RiceCategory{
	{
		Key: "gao_deo_thom", Label: "Gạo dẻo thơm",
		Products: []RiceProduct{
			{Key: "st_25", Label: "ST 25", Category: "gao_deo_thom"},
			{Key: "st_24", Label: "ST 24", Category: "gao_deo_thom"},
			{Key: "st_21", Label: "ST 21", Category: "gao_deo_thom"},
			{Key: "om_18", Label: "OM 18", Category: "gao_deo_thom"},
			{Key: "om_49", Label: "OM 49", Category: "gao_deo_thom"},
			{Key: "om_5451", Label: "OM 5451", Category: "gao_deo_thom"},
			{Key: "dai_thom_8", Label: "Đài Thơm 8", Category: "gao_deo_thom"},
			{Key: "om_6976", Label: "OM 6976", Category: "gao_deo_thom"},
			{Key: "nhat", Label: "Nhật", Category: "gao_deo_thom"},
			{Key: "lien_huong", Label: "Liên Hương", Category: "gao_deo_thom"},
			{Key: "mien", Label: "Miên", Category: "gao_deo_thom"},
			{Key: "dai_loan", Label: "Đài Loan", Category: "gao_deo_thom"},
		},
	},
	{
		Key: "gao_kho", Label: "Gạo khô",
		Products: []RiceProduct{
			{Key: "tai_nguyen", Label: "Tài Nguyên", Category: "gao_kho"},
			{Key: "soc", Label: "Sóc", Category: "gao_kho"},
			{Key: "so_ri", Label: "Sơ Ri", Category: "gao_kho"},
			{Key: "mong_chim", Label: "Móng Chim", Category: "gao_kho"},
			{Key: "ham_chau_sieu", Label: "Hàm Châu siêu", Category: "gao_kho"},
			{Key: "ir_504", Label: "IR 504", Category: "gao_kho"},
			{Key: "q5", Label: "Q5", Category: "gao_kho"},
			{Key: "an_no", Label: "Ấn nở", Category: "gao_kho"},
			{Key: "myanmar", Label: "Myanmar", Category: "gao_kho"},
		},
	},
	{
		Key: "tam_deo_thom", Label: "Tấm dẻo thơm",
		Products: []RiceProduct{
			{Key: "tam_st_25", Label: "Tấm ST 25", Category: "tam_deo_thom"},
			{Key: "tam_st_24", Label: "Tấm ST 24", Category: "tam_deo_thom"},
			{Key: "tam_st_21", Label: "Tấm ST 21", Category: "tam_deo_thom"},
			{Key: "tam_om_18", Label: "Tấm OM 18", Category: "tam_deo_thom"},
			{Key: "tam_om_49", Label: "Tấm OM 49", Category: "tam_deo_thom"},
			{Key: "tam_om_5451", Label: "Tấm OM 5451", Category: "tam_deo_thom"},
			{Key: "tam_dai_thom_8", Label: "Tấm Đài Thơm 8", Category: "tam_deo_thom"},
			{Key: "tam_om_6976", Label: "Tấm OM 6976", Category: "tam_deo_thom"},
			{Key: "tam_nhat", Label: "Tấm Nhật", Category: "tam_deo_thom"},
			{Key: "tam_lien_huong", Label: "Tấm Liên Hương", Category: "tam_deo_thom"},
			{Key: "tam_mien", Label: "Tấm Miên", Category: "tam_deo_thom"},
			{Key: "tam_dai_loan", Label: "Tấm Đài Loan", Category: "tam_deo_thom"},
		},
	},
	{
		Key: "tam_kho", Label: "Tấm khô",
		Products: []RiceProduct{
			{Key: "tam_tai_nguyen", Label: "Tấm Tài Nguyên", Category: "tam_kho"},
			{Key: "tam_soc", Label: "Tấm Sóc", Category: "tam_kho"},
			{Key: "tam_so_ri", Label: "Tấm Sơ Ri", Category: "tam_kho"},
			{Key: "tam_mong_chim", Label: "Tấm Móng Chim", Category: "tam_kho"},
			{Key: "tam_ham_chau_sieu", Label: "Tấm Hàm Châu siêu", Category: "tam_kho"},
			{Key: "tam_ir_504", Label: "Tấm IR 504", Category: "tam_kho"},
			{Key: "tam_q5", Label: "Tấm Q5", Category: "tam_kho"},
			{Key: "tam_an_no", Label: "Tấm Ấn nở", Category: "tam_kho"},
			{Key: "tam_myanmar", Label: "Tấm Myanmar", Category: "tam_kho"},
		},
	},
	{
		Key: "nep", Label: "Nếp",
		Products: []RiceProduct{
			{Key: "sap_moi", Label: "Sáp Mới", Category: "nep"},
			{Key: "sap_cu", Label: "Sáp cũ", Category: "nep"},
			{Key: "nep_la_moi", Label: "Nếp Lá mới", Category: "nep"},
			{Key: "nep_la_cu", Label: "Nếp Lá cũ", Category: "nep"},
			{Key: "bac_hat_lon", Label: "Bắc Hạt Lớn", Category: "nep"},
			{Key: "bac_hat_nho", Label: "Bắc Hạt Nhỏ", Category: "nep"},
			{Key: "nep_thai", Label: "Nếp Thái", Category: "nep"},
			{Key: "nep_than", Label: "Nếp Than", Category: "nep"},
			{Key: "huyet_rong", Label: "Huyết Rồng", Category: "nep"},
		},
	},
}

// AllProductKeys returns a set of all valid product keys.
func AllProductKeys() map[string]bool {
	m := make(map[string]bool)
	for _, cat := range RiceCategories {
		for _, p := range cat.Products {
			m[p.Key] = true
		}
	}
	return m
}

// AllCategoryKeys returns a set of all valid category keys.
func AllCategoryKeys() map[string]bool {
	m := make(map[string]bool)
	for _, cat := range RiceCategories {
		m[cat.Key] = true
	}
	return m
}

// ProductByKey looks up a product by its key.
func ProductByKey(key string) *RiceProduct {
	for _, cat := range RiceCategories {
		for i, p := range cat.Products {
			if p.Key == key {
				return &cat.Products[i]
			}
		}
	}
	return nil
}

// CategoryByKey looks up a category by its key.
func CategoryByKey(key string) *RiceCategory {
	for i, cat := range RiceCategories {
		if cat.Key == key {
			return &RiceCategories[i]
		}
	}
	return nil
}

// ValidateProductInCategory checks if a product key belongs to the given category.
func ValidateProductInCategory(categoryKey, productKey string) bool {
	cat := CategoryByKey(categoryKey)
	if cat == nil {
		return false
	}
	for _, p := range cat.Products {
		if p.Key == productKey {
			return true
		}
	}
	return false
}
