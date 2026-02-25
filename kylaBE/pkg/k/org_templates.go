package k

// TeamTemplate represents a team within a department or branch
type TeamTemplate struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// DepartmentTemplate represents a department within a branch
type DepartmentTemplate struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Teams       []TeamTemplate `json:"teams,omitempty"`
}

// BranchTemplate represents a branch of an organization
type BranchTemplate struct {
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Departments []DepartmentTemplate `json:"departments,omitempty"`
	SubBranches []BranchTemplate     `json:"sub_branches,omitempty"` // Nested sub-branches
	Teams       []TeamTemplate       `json:"teams,omitempty"`        // Teams directly under a branch
}

// OrganizationTemplate represents a template for an organization structure
type OrganizationTemplate struct {
	Industry            string           `json:"industry"`
	Subindustry         string           `json:"subindustry"`
	TemplateName        string           `json:"template_name,omitempty"`
	TemplateDescription string           `json:"template_description,omitempty"`
	Branches            []BranchTemplate `json:"branches"`
}

// OrganizationTemplates is a map of industry-specific organization templates
var OrganizationTemplates = map[string]map[string]OrganizationTemplate{
	"Default": {
		"Generic": {
			Industry:            "Default",
			Subindustry:         "Generic",
			TemplateName:        "Standard Organization Structure",
			TemplateDescription: "A general-purpose organizational structure suitable for most businesses with standard departments and teams.",
			Branches: []BranchTemplate{
				{
					Name:        "Headquarters",
					Description: "Main company headquarters",
					Departments: []DepartmentTemplate{
						{
							Name:        "Administration",
							Description: "Administrative functions and operations",
							Teams: []TeamTemplate{
								{
									Name:        "Office Management",
									Description: "Manages office operations and facilities",
								},
							},
						},
						{
							Name:        "Human Resources",
							Description: "Talent management and employee relations",
						},
						{
							Name:        "Finance",
							Description: "Financial management and accounting",
						},
						{
							Name:        "Operations",
							Description: "Core business operations",
						},
					},
				},
				{
					Name:        "Regional Office",
					Description: "Regional operations center",
					Departments: []DepartmentTemplate{
						{
							Name:        "Local Operations",
							Description: "Regional business activities",
						},
						{
							Name:        "Sales",
							Description: "Regional sales team",
						},
					},
				},
			},
		},
	},
	"Technology": {
		"Software": {
			Industry:            "Technology",
			Subindustry:         "Software",
			TemplateName:        "Software Company Structure",
			TemplateDescription: "An organizational structure for software companies with engineering, product, and support teams.",
			Branches: []BranchTemplate{
				{
					Name:        "Headquarters",
					Description: "Corporate headquarters",
					Departments: []DepartmentTemplate{
						{
							Name:        "Engineering",
							Description: "Software development and engineering",
							Teams: []TeamTemplate{
								{
									Name:        "Frontend",
									Description: "Frontend development team",
								},
								{
									Name:        "Backend",
									Description: "Backend development team",
								},
								{
									Name:        "DevOps",
									Description: "Infrastructure and deployment",
								},
								{
									Name:        "QA",
									Description: "Quality assurance and testing",
								},
							},
						},
						{
							Name:        "Product",
							Description: "Product management and design",
							Teams: []TeamTemplate{
								{
									Name:        "Product Management",
									Description: "Product roadmap and features",
								},
								{
									Name:        "UX/UI Design",
									Description: "User experience and interface design",
								},
							},
						},
						{
							Name:        "Marketing",
							Description: "Marketing and communications",
							Teams: []TeamTemplate{
								{
									Name:        "Digital Marketing",
									Description: "Online marketing and campaigns",
								},
								{
									Name:        "Content",
									Description: "Content creation and management",
								},
							},
						},
						{
							Name:        "Sales",
							Description: "Sales and revenue generation",
							Teams: []TeamTemplate{
								{
									Name:        "Enterprise Sales",
									Description: "Large business accounts",
								},
								{
									Name:        "SMB Sales",
									Description: "Small and medium business accounts",
								},
							},
						},
						{
							Name:        "Customer Success",
							Description: "Customer support and success",
						},
						{
							Name:        "Finance",
							Description: "Financial operations",
						},
						{
							Name:        "HR",
							Description: "Human resources",
						},
					},
				},
				{
					Name:        "Regional Development Center",
					Description: "Regional software development office",
					Departments: []DepartmentTemplate{
						{
							Name:        "Engineering",
							Description: "Regional development team",
							Teams: []TeamTemplate{
								{
									Name:        "Development",
									Description: "Software development team",
								},
								{
									Name:        "QA",
									Description: "Quality assurance team",
								},
							},
						},
						{
							Name:        "Operations",
							Description: "Local operations",
						},
					},
				},
			},
		},
		"Hardware": {
			Industry:            "Technology",
			Subindustry:         "Hardware",
			TemplateName:        "Hardware Company Structure",
			TemplateDescription: "A structure for hardware manufacturers with R&D, manufacturing, and quality assurance teams.",
			Branches: []BranchTemplate{
				{
					Name:        "Headquarters",
					Description: "Corporate headquarters",
					Departments: []DepartmentTemplate{
						{
							Name:        "R&D",
							Description: "Research and development",
							Teams: []TeamTemplate{
								{
									Name:        "Hardware Design",
									Description: "Physical product design",
								},
								{
									Name:        "Firmware",
									Description: "Device firmware development",
								},
								{
									Name:        "Testing",
									Description: "Hardware testing and validation",
								},
							},
						},
						{
							Name:        "Manufacturing",
							Description: "Production oversight",
							Teams: []TeamTemplate{
								{
									Name:        "Supply Chain",
									Description: "Component sourcing and logistics",
								},
								{
									Name:        "Quality Control",
									Description: "Manufacturing quality assurance",
								},
							},
						},
						{
							Name:        "Product Management",
							Description: "Product development and roadmap",
						},
						{
							Name:        "Sales",
							Description: "Sales and distribution",
						},
						{
							Name:        "Marketing",
							Description: "Product marketing",
						},
					},
				},
				{
					Name:        "Manufacturing Plant",
					Description: "Production facility",
					Departments: []DepartmentTemplate{
						{
							Name:        "Production",
							Description: "Product assembly and manufacturing",
							Teams: []TeamTemplate{
								{
									Name:        "Assembly",
									Description: "Product assembly line",
								},
								{
									Name:        "Packaging",
									Description: "Product packaging",
								},
							},
						},
						{
							Name:        "Quality Assurance",
							Description: "Production quality control",
						},
						{
							Name:        "Logistics",
							Description: "Shipping and receiving",
						},
					},
				},
			},
		},
	},
	"Healthcare": {
		"Hospital": {
			Industry:            "Healthcare",
			Subindustry:         "Hospital",
			TemplateName:        "Hospital Organizational Structure",
			TemplateDescription: "A comprehensive structure for hospitals with clinical departments, administrative functions, and support services.",
			Branches: []BranchTemplate{
				{
					Name:        "Main Hospital",
					Description: "Primary care facility",
					Departments: []DepartmentTemplate{
						{
							Name:        "Emergency",
							Description: "Emergency care department",
							Teams: []TeamTemplate{
								{
									Name:        "Trauma Team",
									Description: "Critical emergency response",
								},
								{
									Name:        "Triage",
									Description: "Patient assessment and routing",
								},
							},
						},
						{
							Name:        "Surgery",
							Description: "Surgical procedures",
							Teams: []TeamTemplate{
								{
									Name:        "General Surgery",
									Description: "Common surgical procedures",
								},
								{
									Name:        "Specialized Surgery",
									Description: "Complex surgical procedures",
								},
							},
						},
						{
							Name:        "Internal Medicine",
							Description: "Diagnosis and non-surgical treatment",
							Teams: []TeamTemplate{
								{
									Name:        "Cardiology",
									Description: "Heart health specialists",
								},
								{
									Name:        "Gastroenterology",
									Description: "Digestive system specialists",
								},
							},
						},
						{
							Name:        "Pediatrics",
							Description: "Child healthcare",
						},
						{
							Name:        "Obstetrics & Gynecology",
							Description: "Women's health and childbirth",
						},
						{
							Name:        "Radiology",
							Description: "Imaging and diagnostics",
						},
						{
							Name:        "Laboratory",
							Description: "Medical testing",
						},
						{
							Name:        "Pharmacy",
							Description: "Medication management",
						},
						{
							Name:        "Administration",
							Description: "Hospital management",
							Teams: []TeamTemplate{
								{
									Name:        "Finance",
									Description: "Financial operations",
								},
								{
									Name:        "HR",
									Description: "Staff management",
								},
							},
						},
					},
				},
				{
					Name:        "Outpatient Clinic",
					Description: "Non-emergency care facility",
					Departments: []DepartmentTemplate{
						{
							Name:        "General Practice",
							Description: "Routine medical care",
						},
						{
							Name:        "Specialist Consultations",
							Description: "Specialized medical consultations",
						},
						{
							Name:        "Physical Therapy",
							Description: "Rehabilitation services",
						},
					},
				},
			},
		},
		"Clinic": {
			Industry:            "Healthcare",
			Subindustry:         "Clinic",
			TemplateName:        "Medical Clinic Structure",
			TemplateDescription: "An organizational structure for medical clinics providing outpatient care with specialized departments and administrative support.",
			Branches: []BranchTemplate{
				{
					Name:        "Main Clinic",
					Description: "Primary care location",
					Departments: []DepartmentTemplate{
						{
							Name:        "General Practice",
							Description: "Routine medical care",
							Teams: []TeamTemplate{
								{
									Name:        "Family Medicine",
									Description: "Comprehensive family care",
								},
								{
									Name:        "Pediatrics",
									Description: "Child healthcare",
								},
							},
						},
						{
							Name:        "Specialists",
							Description: "Specialized medical care",
						},
						{
							Name:        "Nursing",
							Description: "Patient care support",
						},
						{
							Name:        "Administration",
							Description: "Clinic management",
						},
					},
				},
			},
		},
	},
	"Education": {
		"University": {
			Industry:            "Education",
			Subindustry:         "University",
			TemplateName:        "University Structure",
			TemplateDescription: "A comprehensive organizational structure for higher education institutions with academic departments, research divisions, and administrative functions.",
			Branches: []BranchTemplate{
				{
					Name:        "Main Campus",
					Description: "Primary university campus",
					Departments: []DepartmentTemplate{
						{
							Name:        "Academic Affairs",
							Description: "Academic programs and policies",
							Teams: []TeamTemplate{
								{
									Name:        "Curriculum Development",
									Description: "Program and course design",
								},
								{
									Name:        "Faculty Affairs",
									Description: "Faculty support and development",
								},
							},
						},
						{
							Name:        "College of Arts & Sciences",
							Description: "Liberal arts programs",
							Teams: []TeamTemplate{
								{
									Name:        "Humanities",
									Description: "Literature, philosophy, etc.",
								},
								{
									Name:        "Natural Sciences",
									Description: "Biology, chemistry, physics, etc.",
								},
								{
									Name:        "Social Sciences",
									Description: "Psychology, sociology, etc.",
								},
							},
						},
						{
							Name:        "College of Engineering",
							Description: "Engineering programs",
						},
						{
							Name:        "Business School",
							Description: "Business and management programs",
						},
						{
							Name:        "Student Affairs",
							Description: "Student services and support",
							Teams: []TeamTemplate{
								{
									Name:        "Admissions",
									Description: "Student recruitment and enrollment",
								},
								{
									Name:        "Financial Aid",
									Description: "Student financial assistance",
								},
								{
									Name:        "Residential Life",
									Description: "Campus housing management",
								},
							},
						},
						{
							Name:        "Research",
							Description: "Research programs and grants",
						},
						{
							Name:        "Administration",
							Description: "University operations",
							Teams: []TeamTemplate{
								{
									Name:        "Finance",
									Description: "Financial management",
								},
								{
									Name:        "HR",
									Description: "Human resources",
								},
								{
									Name:        "Facilities",
									Description: "Campus maintenance and operations",
								},
							},
						},
					},
				},
				{
					Name:        "Satellite Campus",
					Description: "Secondary university location",
					Departments: []DepartmentTemplate{
						{
							Name:        "Academic Programs",
							Description: "Satellite campus programs",
						},
						{
							Name:        "Administration",
							Description: "Campus operations",
						},
						{
							Name:        "Student Services",
							Description: "Student support",
						},
					},
				},
			},
		},
		"K-12 School": {
			Industry:            "Education",
			Subindustry:         "K-12 School",
			TemplateName:        "K-12 School Structure",
			TemplateDescription: "An organizational framework for primary and secondary educational institutions with academic, administrative, and extracurricular departments.",
			Branches: []BranchTemplate{
				{
					Name:        "Main School",
					Description: "Primary school campus",
					Departments: []DepartmentTemplate{
						{
							Name:        "Elementary Division",
							Description: "Grades K-5",
							Teams: []TeamTemplate{
								{
									Name:        "Lower Elementary",
									Description: "Grades K-2",
								},
								{
									Name:        "Upper Elementary",
									Description: "Grades 3-5",
								},
							},
						},
						{
							Name:        "Middle School Division",
							Description: "Grades 6-8",
						},
						{
							Name:        "High School Division",
							Description: "Grades 9-12",
							Teams: []TeamTemplate{
								{
									Name:        "Science Department",
									Description: "Science curriculum",
								},
								{
									Name:        "Mathematics Department",
									Description: "Math curriculum",
								},
								{
									Name:        "English Department",
									Description: "Language arts curriculum",
								},
							},
						},
						{
							Name:        "Special Education",
							Description: "Support for students with special needs",
						},
						{
							Name:        "Athletics",
							Description: "Sports programs",
						},
						{
							Name:        "Arts",
							Description: "Visual and performing arts",
						},
						{
							Name:        "Administration",
							Description: "School management",
							Teams: []TeamTemplate{
								{
									Name:        "Principal's Office",
									Description: "School leadership",
								},
								{
									Name:        "Counseling",
									Description: "Student guidance and support",
								},
							},
						},
					},
				},
			},
		},
	},
	"Manufacturing": {
		"Automotive": {
			Industry:            "Manufacturing",
			Subindustry:         "Automotive",
			TemplateName:        "Automotive Manufacturing Structure",
			TemplateDescription: "A specialized structure for automotive manufacturers with design, production, and quality control departments.",
			Branches: []BranchTemplate{
				{
					Name:        "Headquarters",
					Description: "Corporate headquarters",
					Departments: []DepartmentTemplate{
						{
							Name:        "Executive",
							Description: "Executive leadership",
						},
						{
							Name:        "Engineering",
							Description: "Product design and engineering",
							Teams: []TeamTemplate{
								{
									Name:        "Vehicle Design",
									Description: "Overall vehicle design",
								},
								{
									Name:        "Powertrain",
									Description: "Engine and transmission systems",
								},
								{
									Name:        "Electronics",
									Description: "Vehicle electronics and software",
								},
							},
						},
						{
							Name:        "Product Development",
							Description: "New product planning",
						},
						{
							Name:        "Marketing",
							Description: "Brand and product marketing",
						},
						{
							Name:        "Sales",
							Description: "Sales strategy and dealer relations",
						},
						{
							Name:        "Finance",
							Description: "Financial operations",
						},
						{
							Name:        "HR",
							Description: "Human resources",
						},
					},
				},
				{
					Name:        "Manufacturing Plant",
					Description: "Vehicle production facility",
					Departments: []DepartmentTemplate{
						{
							Name:        "Production",
							Description: "Vehicle assembly",
							Teams: []TeamTemplate{
								{
									Name:        "Body Shop",
									Description: "Vehicle body assembly",
								},
								{
									Name:        "Paint Shop",
									Description: "Vehicle painting",
								},
								{
									Name:        "Final Assembly",
									Description: "Vehicle completion",
								},
							},
						},
						{
							Name:        "Quality Control",
							Description: "Product quality assurance",
						},
						{
							Name:        "Logistics",
							Description: "Supply chain and shipping",
						},
						{
							Name:        "Maintenance",
							Description: "Equipment maintenance",
						},
						{
							Name:        "Plant Administration",
							Description: "Facility management",
						},
					},
				},
				{
					Name:        "R&D Center",
					Description: "Research and development facility",
					Departments: []DepartmentTemplate{
						{
							Name:        "Research",
							Description: "Advanced technology research",
							Teams: []TeamTemplate{
								{
									Name:        "EV Technology",
									Description: "Electric vehicle research",
								},
								{
									Name:        "Autonomous Driving",
									Description: "Self-driving technology",
								},
							},
						},
						{
							Name:        "Testing",
							Description: "Product testing and validation",
						},
					},
				},
			},
		},
		"Consumer Goods": {
			Industry:            "Manufacturing",
			Subindustry:         "Consumer Goods",
			TemplateName:        "Consumer Goods Manufacturing Structure",
			TemplateDescription: "An organizational structure for consumer products manufacturers with product development, production, and marketing departments.",
			Branches: []BranchTemplate{
				{
					Name:        "Headquarters",
					Description: "Corporate headquarters",
					Departments: []DepartmentTemplate{
						{
							Name:        "Product Development",
							Description: "New product creation",
							Teams: []TeamTemplate{
								{
									Name:        "Design",
									Description: "Product design",
								},
								{
									Name:        "R&D",
									Description: "Research and development",
								},
							},
						},
						{
							Name:        "Marketing",
							Description: "Brand and product marketing",
						},
						{
							Name:        "Sales",
							Description: "Retail and distribution relationships",
						},
						{
							Name:        "Supply Chain",
							Description: "Supply chain management",
						},
						{
							Name:        "Finance",
							Description: "Financial operations",
						},
					},
				},
				{
					Name:        "Production Facility",
					Description: "Manufacturing plant",
					Departments: []DepartmentTemplate{
						{
							Name:        "Manufacturing",
							Description: "Product production",
							Teams: []TeamTemplate{
								{
									Name:        "Production Line A",
									Description: "Primary production line",
								},
								{
									Name:        "Production Line B",
									Description: "Secondary production line",
								},
							},
						},
						{
							Name:        "Quality Control",
							Description: "Product quality assurance",
						},
						{
							Name:        "Warehouse",
							Description: "Inventory management",
						},
					},
				},
			},
		},
	},
	"Financial Services": {
		"Banking": {
			Industry:            "Financial Services",
			Subindustry:         "Banking",
			TemplateName:        "Banking Organization Structure",
			TemplateDescription: "A comprehensive structure for banking institutions with retail, commercial, investment divisions, and regulatory compliance functions.",
			Branches: []BranchTemplate{
				{
					Name:        "Headquarters",
					Description: "Corporate headquarters",
					Departments: []DepartmentTemplate{
						{
							Name:        "Retail Banking",
							Description: "Consumer banking services",
							Teams: []TeamTemplate{
								{
									Name:        "Checking & Savings",
									Description: "Deposit accounts",
								},
								{
									Name:        "Mortgage",
									Description: "Home loans",
								},
								{
									Name:        "Consumer Lending",
									Description: "Personal loans and credit",
								},
							},
						},
						{
							Name:        "Commercial Banking",
							Description: "Business banking services",
							Teams: []TeamTemplate{
								{
									Name:        "Small Business",
									Description: "Small business services",
								},
								{
									Name:        "Corporate Banking",
									Description: "Large business services",
								},
							},
						},
						{
							Name:        "Investment Banking",
							Description: "Capital markets and advisory",
						},
						{
							Name:        "Wealth Management",
							Description: "High-net-worth client services",
						},
						{
							Name:        "Risk & Compliance",
							Description: "Risk management and regulatory compliance",
							Teams: []TeamTemplate{
								{
									Name:        "Credit Risk",
									Description: "Lending risk assessment",
								},
								{
									Name:        "Compliance",
									Description: "Regulatory compliance",
								},
							},
						},
						{
							Name:        "IT",
							Description: "Information technology",
							Teams: []TeamTemplate{
								{
									Name:        "Core Banking Systems",
									Description: "Primary banking platforms",
								},
								{
									Name:        "Digital Banking",
									Description: "Online and mobile banking",
								},
								{
									Name:        "Cybersecurity",
									Description: "Information security",
								},
							},
						},
						{
							Name:        "Operations",
							Description: "Banking operations",
						},
						{
							Name:        "Finance",
							Description: "Financial management",
						},
						{
							Name:        "HR",
							Description: "Human resources",
						},
					},
				},
				{
					Name:        "Regional Branch",
					Description: "Customer-facing banking office",
					Departments: []DepartmentTemplate{
						{
							Name:        "Customer Service",
							Description: "Front-line customer support",
						},
						{
							Name:        "Personal Banking",
							Description: "Individual account services",
						},
						{
							Name:        "Business Banking",
							Description: "Local business services",
						},
						{
							Name:        "Branch Operations",
							Description: "Branch management",
						},
					},
				},
			},
		},
		"Insurance": {
			Industry:            "Financial Services",
			Subindustry:         "Insurance",
			TemplateName:        "Insurance Company Structure",
			TemplateDescription: "An organizational framework for insurance providers with underwriting, claims processing, and customer service departments.",
			Branches: []BranchTemplate{
				{
					Name:        "Headquarters",
					Description: "Corporate headquarters",
					Departments: []DepartmentTemplate{
						{
							Name:        "Underwriting",
							Description: "Risk assessment and policy pricing",
							Teams: []TeamTemplate{
								{
									Name:        "Personal Lines",
									Description: "Individual insurance policies",
								},
								{
									Name:        "Commercial Lines",
									Description: "Business insurance policies",
								},
							},
						},
						{
							Name:        "Claims",
							Description: "Claims processing and payment",
							Teams: []TeamTemplate{
								{
									Name:        "Claims Intake",
									Description: "Initial claims processing",
								},
								{
									Name:        "Claims Investigation",
									Description: "Claims verification",
								},
								{
									Name:        "Claims Payment",
									Description: "Benefit disbursement",
								},
							},
						},
						{
							Name:        "Actuarial",
							Description: "Statistical risk assessment",
						},
						{
							Name:        "Product Development",
							Description: "Insurance product design",
						},
						{
							Name:        "Sales",
							Description: "Agent relations and sales",
						},
						{
							Name:        "Marketing",
							Description: "Brand and product marketing",
						},
						{
							Name:        "Legal & Compliance",
							Description: "Legal matters and regulatory compliance",
						},
						{
							Name:        "IT",
							Description: "Information technology",
						},
						{
							Name:        "Finance",
							Description: "Financial operations",
						},
					},
				},
				{
					Name:        "Regional Office",
					Description: "Regional operations center",
					Departments: []DepartmentTemplate{
						{
							Name:        "Sales",
							Description: "Regional sales team",
						},
						{
							Name:        "Claims Processing",
							Description: "Regional claims handling",
						},
						{
							Name:        "Customer Service",
							Description: "Policy service and support",
						},
					},
				},
			},
		},
	},
	"Retail": {
		"Department Store": {
			Industry:            "Retail",
			Subindustry:         "Department Store",
			TemplateName:        "Department Store Structure",
			TemplateDescription: "An organizational structure for multi-department retail stores with merchandise categories, customer service, and operations teams.",
			Branches: []BranchTemplate{
				{
					Name:        "Corporate Office",
					Description: "Corporate headquarters",
					Departments: []DepartmentTemplate{
						{
							Name:        "Merchandising",
							Description: "Product selection and buying",
							Teams: []TeamTemplate{
								{
									Name:        "Apparel",
									Description: "Clothing and accessories buying",
								},
								{
									Name:        "Home Goods",
									Description: "Home products buying",
								},
								{
									Name:        "Electronics",
									Description: "Electronics buying",
								},
							},
						},
						{
							Name:        "Marketing",
							Description: "Brand and promotional marketing",
							Teams: []TeamTemplate{
								{
									Name:        "Advertising",
									Description: "Promotional campaigns",
								},
								{
									Name:        "Digital Marketing",
									Description: "Online marketing",
								},
							},
						},
						{
							Name:        "Store Operations",
							Description: "Retail location management",
						},
						{
							Name:        "E-commerce",
							Description: "Online retail operations",
							Teams: []TeamTemplate{
								{
									Name:        "Website Management",
									Description: "Online store maintenance",
								},
								{
									Name:        "Fulfillment",
									Description: "Order processing and shipping",
								},
							},
						},
						{
							Name:        "Supply Chain",
							Description: "Inventory and logistics",
						},
						{
							Name:        "Finance",
							Description: "Financial operations",
						},
						{
							Name:        "HR",
							Description: "Human resources",
						},
					},
				},
				{
					Name:        "Flagship Store",
					Description: "Primary retail location",
					Departments: []DepartmentTemplate{
						{
							Name:        "Sales Floor",
							Description: "Customer-facing sales area",
							Teams: []TeamTemplate{
								{
									Name:        "Women's Department",
									Description: "Women's clothing and accessories",
								},
								{
									Name:        "Men's Department",
									Description: "Men's clothing and accessories",
								},
								{
									Name:        "Home Department",
									Description: "Home goods and furnishings",
								},
								{
									Name:        "Beauty Department",
									Description: "Cosmetics and skincare",
								},
							},
						},
						{
							Name:        "Customer Service",
							Description: "Returns and customer assistance",
						},
						{
							Name:        "Inventory Management",
							Description: "Stock management and replenishment",
						},
						{
							Name:        "Store Administration",
							Description: "Store operations management",
						},
					},
				},
				{
					Name:        "Distribution Center",
					Description: "Inventory storage and distribution",
					Departments: []DepartmentTemplate{
						{
							Name:        "Receiving",
							Description: "Incoming inventory processing",
						},
						{
							Name:        "Warehousing",
							Description: "Inventory storage",
						},
						{
							Name:        "Shipping",
							Description: "Order fulfillment and shipping",
						},
						{
							Name:        "Operations",
							Description: "Facility management",
						},
					},
				},
			},
		},
		"Grocery": {
			Industry:            "Retail",
			Subindustry:         "Grocery",
			TemplateName:        "Grocery Store Structure",
			TemplateDescription: "An organizational structure for grocery retailers with fresh foods, dry goods, and customer service departments.",
			Branches: []BranchTemplate{
				{
					Name:        "Headquarters",
					Description: "Corporate headquarters",
					Departments: []DepartmentTemplate{
						{
							Name:        "Merchandising",
							Description: "Product selection and procurement",
							Teams: []TeamTemplate{
								{
									Name:        "Produce",
									Description: "Fresh fruits and vegetables",
								},
								{
									Name:        "Meat & Seafood",
									Description: "Meat and seafood products",
								},
								{
									Name:        "Grocery",
									Description: "Packaged goods",
								},
							},
						},
						{
							Name:        "Operations",
							Description: "Store operations management",
						},
						{
							Name:        "Marketing",
							Description: "Promotions and advertising",
						},
						{
							Name:        "Supply Chain",
							Description: "Distribution and logistics",
						},
						{
							Name:        "Finance",
							Description: "Financial operations",
						},
						{
							Name:        "HR",
							Description: "Human resources",
						},
					},
				},
				{
					Name:        "Store",
					Description: "Retail grocery location",
					Departments: []DepartmentTemplate{
						{
							Name:        "Front End",
							Description: "Checkout and customer service",
						},
						{
							Name:        "Produce",
							Description: "Fresh fruits and vegetables",
						},
						{
							Name:        "Meat & Seafood",
							Description: "Fresh meat and seafood",
						},
						{
							Name:        "Bakery",
							Description: "Fresh baked goods",
						},
						{
							Name:        "Deli",
							Description: "Prepared foods and specialty items",
						},
						{
							Name:        "Grocery",
							Description: "Packaged foods and household items",
						},
						{
							Name:        "Dairy & Frozen",
							Description: "Refrigerated and frozen products",
						},
						{
							Name:        "Store Management",
							Description: "Store operations and administration",
						},
					},
				},
			},
		},
	},
	"Construction": {
		"General Contracting": {
			Industry:            "Construction",
			Subindustry:         "General Contracting",
			TemplateName:        "General Contracting Structure",
			TemplateDescription: "An organizational framework for construction companies with project management, estimating, and field operations teams.",
			Branches: []BranchTemplate{
				{
					Name:        "Headquarters",
					Description: "Corporate headquarters",
					Departments: []DepartmentTemplate{
						{
							Name:        "Estimating",
							Description: "Project cost analysis and bidding",
						},
						{
							Name:        "Project Management",
							Description: "Construction project oversight",
							Teams: []TeamTemplate{
								{
									Name:        "Residential Projects",
									Description: "Home construction management",
								},
								{
									Name:        "Commercial Projects",
									Description: "Business construction management",
								},
							},
						},
						{
							Name:        "Engineering",
							Description: "Construction engineering",
						},
						{
							Name:        "Operations",
							Description: "Field operations management",
						},
						{
							Name:        "Procurement",
							Description: "Materials and subcontractor sourcing",
						},
						{
							Name:        "Safety & Compliance",
							Description: "Safety protocols and regulatory compliance",
						},
						{
							Name:        "Finance",
							Description: "Financial operations",
						},
						{
							Name:        "Business Development",
							Description: "Client acquisition and relationships",
						},
					},
				},
				{
					Name:        "Regional Office",
					Description: "Regional operations center",
					Departments: []DepartmentTemplate{
						{
							Name:        "Project Management",
							Description: "Regional project oversight",
						},
						{
							Name:        "Field Operations",
							Description: "Construction site management",
						},
						{
							Name:        "Administration",
							Description: "Office management",
						},
					},
				},
				{
					Name:        "Construction Site",
					Description: "Active project location",
					Departments: []DepartmentTemplate{
						{
							Name:        "Site Management",
							Description: "On-site project leadership",
						},
						{
							Name:        "Construction Crew",
							Description: "Construction workers",
							Teams: []TeamTemplate{
								{
									Name:        "Structural Team",
									Description: "Framing and structural work",
								},
								{
									Name:        "Finishing Team",
									Description: "Interior and exterior finishing",
								},
							},
						},
						{
							Name:        "Safety Oversight",
							Description: "On-site safety management",
						},
					},
				},
			},
		},
		"Specialty Contracting": {
			Industry:            "Construction",
			Subindustry:         "Specialty Contracting",
			TemplateName:        "Specialty Contracting Structure",
			TemplateDescription: "An organizational structure for specialized contractors (electrical, plumbing, HVAC) with technical teams and project management.",
			Branches: []BranchTemplate{
				{
					Name:        "Headquarters",
					Description: "Company headquarters",
					Departments: []DepartmentTemplate{
						{
							Name:        "Estimating",
							Description: "Project cost analysis",
						},
						{
							Name:        "Project Management",
							Description: "Project oversight",
						},
						{
							Name:        "Field Operations",
							Description: "On-site work management",
						},
						{
							Name:        "Administration",
							Description: "Office management and support",
						},
						{
							Name:        "Sales",
							Description: "Business development",
						},
					},
				},
				{
					Name:        "Field Office",
					Description: "Project support location",
					Departments: []DepartmentTemplate{
						{
							Name:        "Project Coordination",
							Description: "Local project management",
						},
						{
							Name:        "Field Crew",
							Description: "Specialized workers",
						},
					},
				},
			},
		},
	},
	"Hospitality": {
		"Hotel": {
			Industry:            "Hospitality",
			Subindustry:         "Hotel",
			TemplateName:        "Hotel Management Structure",
			TemplateDescription: "An organizational framework for hotels with front office, housekeeping, food & beverage, and administrative departments.",
			Branches: []BranchTemplate{
				{
					Name:        "Corporate Office",
					Description: "Corporate headquarters",
					Departments: []DepartmentTemplate{
						{
							Name:        "Operations",
							Description: "Hotel operations management",
						},
						{
							Name:        "Sales & Marketing",
							Description: "Revenue generation and promotion",
							Teams: []TeamTemplate{
								{
									Name:        "Group Sales",
									Description: "Conference and event sales",
								},
								{
									Name:        "Marketing",
									Description: "Brand and promotional marketing",
								},
							},
						},
						{
							Name:        "Revenue Management",
							Description: "Pricing and inventory optimization",
						},
						{
							Name:        "Development",
							Description: "Property development and acquisition",
						},
						{
							Name:        "Finance",
							Description: "Financial operations",
						},
						{
							Name:        "HR",
							Description: "Human resources",
						},
						{
							Name:        "IT",
							Description: "Information technology",
						},
					},
				},
				{
					Name:        "Hotel Property",
					Description: "Individual hotel location",
					Departments: []DepartmentTemplate{
						{
							Name:        "Front Office",
							Description: "Guest reception and services",
							Teams: []TeamTemplate{
								{
									Name:        "Front Desk",
									Description: "Check-in and guest services",
								},
								{
									Name:        "Concierge",
									Description: "Guest assistance and information",
								},
							},
						},
						{
							Name:        "Housekeeping",
							Description: "Room and facility cleaning",
						},
						{
							Name:        "Food & Beverage",
							Description: "Restaurants and catering",
							Teams: []TeamTemplate{
								{
									Name:        "Restaurant",
									Description: "Hotel dining",
								},
								{
									Name:        "Banquets",
									Description: "Event catering",
								},
								{
									Name:        "Room Service",
									Description: "In-room dining",
								},
							},
						},
						{
							Name:        "Engineering & Maintenance",
							Description: "Facility maintenance",
						},
						{
							Name:        "Security",
							Description: "Property and guest safety",
						},
						{
							Name:        "Sales",
							Description: "Local sales team",
						},
						{
							Name:        "Hotel Management",
							Description: "Property leadership",
						},
					},
				},
			},
		},
		"Restaurant": {
			Industry:            "Hospitality",
			Subindustry:         "Restaurant",
			TemplateName:        "Restaurant Management Structure",
			TemplateDescription: "An organizational structure for restaurants with kitchen, front-of-house, and administrative functions.",
			Branches: []BranchTemplate{
				{
					Name:        "Headquarters",
					Description: "Corporate headquarters",
					Departments: []DepartmentTemplate{
						{
							Name:        "Operations",
							Description: "Restaurant operations management",
						},
						{
							Name:        "Culinary",
							Description: "Menu development and food standards",
						},
						{
							Name:        "Marketing",
							Description: "Brand and promotional marketing",
						},
						{
							Name:        "Supply Chain",
							Description: "Food and equipment procurement",
						},
						{
							Name:        "Finance",
							Description: "Financial operations",
						},
						{
							Name:        "HR",
							Description: "Human resources",
						},
						{
							Name:        "Development",
							Description: "New location planning",
						},
					},
				},
				{
					Name:        "Restaurant Location",
					Description: "Individual restaurant",
					Departments: []DepartmentTemplate{
						{
							Name:        "Front of House",
							Description: "Customer-facing operations",
							Teams: []TeamTemplate{
								{
									Name:        "Hosts",
									Description: "Guest greeting and seating",
								},
								{
									Name:        "Servers",
									Description: "Table service",
								},
								{
									Name:        "Bar",
									Description: "Beverage service",
								},
							},
						},
						{
							Name:        "Back of House",
							Description: "Kitchen operations",
							Teams: []TeamTemplate{
								{
									Name:        "Line Cooks",
									Description: "Food preparation",
								},
								{
									Name:        "Dishwashers",
									Description: "Kitchen sanitation",
								},
							},
						},
						{
							Name:        "Management",
							Description: "Restaurant leadership",
						},
					},
				},
			},
		},
	},
	"Government": {
		"Municipal": {
			Industry:            "Government",
			Subindustry:         "Municipal",
			TemplateName:        "Municipal Government Structure",
			TemplateDescription: "An organizational structure for city governments with administrative departments, public works, and community services.",
			Branches: []BranchTemplate{
				{
					Name:        "City Hall",
					Description: "Primary government offices",
					Departments: []DepartmentTemplate{
						{
							Name:        "Mayor's Office",
							Description: "Executive leadership",
						},
						{
							Name:        "City Council",
							Description: "Legislative body",
						},
						{
							Name:        "Finance",
							Description: "Budget and financial management",
							Teams: []TeamTemplate{
								{
									Name:        "Accounting",
									Description: "Financial record-keeping",
								},
								{
									Name:        "Tax Collection",
									Description: "Revenue collection",
								},
							},
						},
						{
							Name:        "Planning & Development",
							Description: "Urban planning and permits",
						},
						{
							Name:        "Public Works",
							Description: "Infrastructure management",
							Teams: []TeamTemplate{
								{
									Name:        "Roads & Transportation",
									Description: "Street maintenance",
								},
								{
									Name:        "Parks & Recreation",
									Description: "Public spaces management",
								},
								{
									Name:        "Water & Sewer",
									Description: "Utility management",
								},
							},
						},
						{
							Name:        "HR",
							Description: "Human resources",
						},
						{
							Name:        "Legal",
							Description: "City attorney's office",
						},
						{
							Name:        "IT",
							Description: "Information technology",
						},
					},
				},
				{
					Name:        "Public Safety",
					Description: "Emergency services",
					Departments: []DepartmentTemplate{
						{
							Name:        "Police Department",
							Description: "Law enforcement",
							Teams: []TeamTemplate{
								{
									Name:        "Patrol",
									Description: "Regular police patrols",
								},
								{
									Name:        "Investigations",
									Description: "Detective bureau",
								},
								{
									Name:        "Administration",
									Description: "Department management",
								},
							},
						},
						{
							Name:        "Fire Department",
							Description: "Fire and rescue services",
							Teams: []TeamTemplate{
								{
									Name:        "Fire Suppression",
									Description: "Firefighting",
								},
								{
									Name:        "EMS",
									Description: "Emergency medical services",
								},
							},
						},
						{
							Name:        "Emergency Management",
							Description: "Disaster preparedness and response",
						},
					},
				},
				{
					Name:        "Public Services",
					Description: "Community services",
					Departments: []DepartmentTemplate{
						{
							Name:        "Library",
							Description: "Public library services",
						},
						{
							Name:        "Parks & Recreation",
							Description: "Public recreation facilities",
						},
						{
							Name:        "Community Development",
							Description: "Social services and programs",
						},
					},
				},
			},
		},
		"Federal": {
			Industry:            "Government",
			Subindustry:         "Federal",
			TemplateName:        "Federal Agency Structure",
			TemplateDescription: "An organizational framework for federal government agencies with administrative, program, and regional offices.",
			Branches: []BranchTemplate{
				{
					Name:        "Headquarters",
					Description: "Agency headquarters",
					Departments: []DepartmentTemplate{
						{
							Name:        "Executive Office",
							Description: "Agency leadership",
						},
						{
							Name:        "Policy & Planning",
							Description: "Policy development",
							Teams: []TeamTemplate{
								{
									Name:        "Strategic Planning",
									Description: "Long-term planning",
								},
								{
									Name:        "Policy Analysis",
									Description: "Policy research and development",
								},
							},
						},
						{
							Name:        "Operations",
							Description: "Program implementation",
						},
						{
							Name:        "Administration",
							Description: "Agency management",
							Teams: []TeamTemplate{
								{
									Name:        "HR",
									Description: "Human resources",
								},
								{
									Name:        "Finance",
									Description: "Budget and financial management",
								},
								{
									Name:        "IT",
									Description: "Information technology",
								},
							},
						},
						{
							Name:        "Legal Counsel",
							Description: "Legal affairs",
						},
						{
							Name:        "Public Affairs",
							Description: "Communications and media relations",
						},
						{
							Name:        "Inspector General",
							Description: "Oversight and compliance",
						},
					},
				},
				{
					Name:        "Regional Office",
					Description: "Regional agency operations",
					Departments: []DepartmentTemplate{
						{
							Name:        "Program Management",
							Description: "Regional program implementation",
						},
						{
							Name:        "Administration",
							Description: "Office management",
						},
						{
							Name:        "Stakeholder Relations",
							Description: "Community and partner engagement",
						},
					},
				},
			},
		},
	},
	"Non-profit": {
		"Charity": {
			Industry:            "Non-profit",
			Subindustry:         "Charity",
			TemplateName:        "Charity Organization Structure",
			TemplateDescription: "An organizational structure for charitable organizations with program delivery, fundraising, and volunteer management departments.",
			Branches: []BranchTemplate{
				{
					Name:        "Headquarters",
					Description: "Organization headquarters",
					Departments: []DepartmentTemplate{
						{
							Name:        "Executive",
							Description: "Leadership and governance",
							Teams: []TeamTemplate{
								{
									Name:        "Executive Office",
									Description: "CEO and executive leadership",
								},
								{
									Name:        "Board Relations",
									Description: "Board of directors support",
								},
							},
						},
						{
							Name:        "Programs",
							Description: "Mission-related activities",
							Teams: []TeamTemplate{
								{
									Name:        "Program Development",
									Description: "Program design and planning",
								},
								{
									Name:        "Program Implementation",
									Description: "Program delivery",
								},
								{
									Name:        "Monitoring & Evaluation",
									Description: "Impact assessment",
								},
							},
						},
						{
							Name:        "Fundraising & Development",
							Description: "Resource mobilization",
							Teams: []TeamTemplate{
								{
									Name:        "Individual Giving",
									Description: "Individual donor relations",
								},
								{
									Name:        "Grants",
									Description: "Foundation and government grants",
								},
								{
									Name:        "Corporate Partnerships",
									Description: "Business relationships",
								},
							},
						},
						{
							Name:        "Marketing & Communications",
							Description: "Public outreach and messaging",
						},
						{
							Name:        "Finance",
							Description: "Financial management",
						},
						{
							Name:        "HR",
							Description: "Human resources",
						},
						{
							Name:        "Volunteer Management",
							Description: "Volunteer coordination",
						},
					},
				},
				{
					Name:        "Regional Office",
					Description: "Regional operations",
					Departments: []DepartmentTemplate{
						{
							Name:        "Program Delivery",
							Description: "Local program implementation",
						},
						{
							Name:        "Community Engagement",
							Description: "Local partnerships and outreach",
						},
						{
							Name:        "Volunteer Coordination",
							Description: "Local volunteer management",
						},
						{
							Name:        "Administration",
							Description: "Office management",
						},
					},
				},
			},
		},
		"Foundation": {
			Industry:            "Non-profit",
			Subindustry:         "Foundation",
			TemplateName:        "Foundation Structure",
			TemplateDescription: "An organizational framework for foundations with grant-making, program management, and investment departments.",
			Branches: []BranchTemplate{
				{
					Name:        "Headquarters",
					Description: "Foundation headquarters",
					Departments: []DepartmentTemplate{
						{
							Name:        "Executive",
							Description: "Leadership and governance",
						},
						{
							Name:        "Grantmaking",
							Description: "Grant review and awarding",
							Teams: []TeamTemplate{
								{
									Name:        "Program Officers",
									Description: "Grant program management",
								},
								{
									Name:        "Grant Administration",
									Description: "Grant processing and oversight",
								},
							},
						},
						{
							Name:        "Investment",
							Description: "Endowment management",
						},
						{
							Name:        "Programs",
							Description: "Direct foundation initiatives",
						},
						{
							Name:        "Communications",
							Description: "Public relations and reporting",
						},
						{
							Name:        "Finance & Administration",
							Description: "Operations management",
						},
						{
							Name:        "Research & Evaluation",
							Description: "Program analysis and learning",
						},
					},
				},
			},
		},
	},
}
