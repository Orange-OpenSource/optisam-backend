package rbac

default allow = false

# Allow admins to do anything.
allow {
	roles["Admin"][input.role]
}

# Normal Users
allow {
  user_apis[input.api]
  input.role = "User"
}

roles := {"Admin":{"SuperAdmin","Admin"},"Normal":{"User"}}
user_apis := {"/optisam.products.v1.ProductService/ListProducts","/optisam.products.v1.ProductService/ListEditors","/optisam.products.v1.ProductService/GetProductDetail",
"/optisam.products.v1.ProductService/GetProductOptions","/optisam.products.v1.ProductService/ListProductAggregationProductView",
"/optisam.products.v1.ProductService/ProductAggregationProductViewOptions","/optisam.products.v1.ProductService/ListProductAggregationView",
"/optisam.products.v1.ProductService/ProductAggregationProductViewDetails","/optisam.products.v1.ProductService/ListEditorProducts","/optisam.products.v1.ProductService/DashboardOverview","/optisam.products.v1.ProductService/ProductsPerEditor",
"/optisam.products.v1.ProductService/ProductsPerMetricType","/optisam.products.v1.ProductService/ComplianceAlert","/optisam.products.v1.ProductService/CounterfeitedProducts","/optisam.products.v1.ProductService/OverdeployedProducts",
"/optisam.products.v1.ProductService/ListAcqRights","/optisam.products.v1.ProductService/ListAcqRightsAggregation","/optisam.products.v1.ProductService/ListAcqRightsAggregationRecords","/optisam.products.v1.ProductService/ListAcqRightsEditors",
"/optisam.products.v1.ProductService/ListAcqRightsProducts","/optisam.products.v1.ProductService/ListAcqRightsMetrics","/optisam.products.v1.ProductService/ListProductAggregation","/optisam.products.v1.ProductService/OverviewProductQuality", "/optisam.products.v1.ProductService/DashboardQualityProducts"
}