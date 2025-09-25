package metadata

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

// MetadataParser handles parsing of RESO metadata XML
type MetadataParser struct {
	Entities map[string]*EntityInfo
	Enums    map[string]*EnumInfo
}

// EntityInfo represents an entity from the metadata
type EntityInfo struct {
	Name        string
	Properties  map[string]*PropertyInfo
	Description string
	IsBaseType  bool
	BaseType    string
}

// PropertyInfo represents a property/field from the metadata
type PropertyInfo struct {
	Name         string
	Type         string
	Description  string
	IsRequired   bool
	IsCollection bool
	EnumType     string
}

// EnumInfo represents an enum type from the metadata
type EnumInfo struct {
	Name        string
	Description string
	Members     map[string]*EnumMemberInfo
}

// EnumMemberInfo represents an enum member/value
type EnumMemberInfo struct {
	Name         string
	Value        string
	StandardName string
	Description  string
}

// EdmxDocument represents the root EDMX document structure
type EdmxDocument struct {
	XMLName      xml.Name     `xml:"Edmx"`
	DataServices DataServices `xml:"DataServices"`
}

// DataServices contains the schema definitions
type DataServices struct {
	Schemas []Schema `xml:"Schema"`
}

// Schema represents a namespace schema
type Schema struct {
	Namespace    string        `xml:"Namespace,attr"`
	EntityTypes  []EntityType  `xml:"EntityType"`
	EnumTypes    []EnumType    `xml:"EnumType"`
	ComplexTypes []ComplexType `xml:"ComplexType"`
}

// EntityType represents an entity definition
type EntityType struct {
	Name       string     `xml:"Name,attr"`
	BaseType   string     `xml:"BaseType,attr"`
	Properties []Property `xml:"Property"`
	Keys       []Key      `xml:"Key"`
}

// Property represents a property/field definition
type Property struct {
	Name      string `xml:"Name,attr"`
	Type      string `xml:"Type,attr"`
	Nullable  string `xml:"Nullable,attr"`
	Scale     string `xml:"Scale,attr"`
	Precision string `xml:"Precision,attr"`
}

// Key represents entity key definition
type Key struct {
	PropertyRefs []PropertyRef `xml:"PropertyRef"`
}

// PropertyRef represents a key property reference
type PropertyRef struct {
	Name string `xml:"Name,attr"`
}

// EnumType represents an enum definition
type EnumType struct {
	Name           string       `xml:"Name,attr"`
	UnderlyingType string       `xml:"UnderlyingType,attr"`
	IsFlags        string       `xml:"IsFlags,attr"`
	Members        []EnumMember `xml:"Member"`
}

// EnumMember represents an enum member/value
type EnumMember struct {
	Name        string       `xml:"Name,attr"`
	Value       string       `xml:"Value,attr"`
	Annotations []Annotation `xml:"Annotation"`
}

// Annotation represents metadata annotations
type Annotation struct {
	Term   string `xml:"Term,attr"`
	String string `xml:"String,attr"`
}

// ComplexType represents complex type definitions
type ComplexType struct {
	Name       string     `xml:"Name,attr"`
	Properties []Property `xml:"Property"`
}

// NewMetadataParser creates a new metadata parser
func NewMetadataParser() *MetadataParser {
	return &MetadataParser{
		Entities: make(map[string]*EntityInfo),
		Enums:    make(map[string]*EnumInfo),
	}
}

// ParseFromFile parses metadata from an XML file
func (p *MetadataParser) ParseFromFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open metadata file: %w", err)
	}
	defer file.Close()

	return p.ParseFromReader(file)
}

// ParseFromReader parses metadata from an XML reader
func (p *MetadataParser) ParseFromReader(reader io.Reader) error {
	var doc EdmxDocument
	decoder := xml.NewDecoder(reader)

	if err := decoder.Decode(&doc); err != nil {
		return fmt.Errorf("failed to parse XML: %w", err)
	}

	// Process all schemas
	for _, schema := range doc.DataServices.Schemas {
		// Parse enum types first (needed for entity properties)
		for _, enumType := range schema.EnumTypes {
			p.parseEnumType(enumType, schema.Namespace)
		}

		// Parse entity types
		for _, entityType := range schema.EntityTypes {
			p.parseEntityType(entityType, schema.Namespace)
		}
	}

	return nil
}

// parseEnumType processes an enum type definition
func (p *MetadataParser) parseEnumType(enumType EnumType, namespace string) {
	fullName := enumType.Name
	if namespace != "" && !strings.Contains(enumType.Name, ".") {
		fullName = namespace + "." + enumType.Name
	}

	enumInfo := &EnumInfo{
		Name:    enumType.Name,
		Members: make(map[string]*EnumMemberInfo),
	}

	// Process enum members
	for _, member := range enumType.Members {
		memberInfo := &EnumMemberInfo{
			Name:  member.Name,
			Value: member.Value,
		}

		// Extract standard name from annotations
		for _, annotation := range member.Annotations {
			if strings.Contains(annotation.Term, "StandardName") {
				memberInfo.StandardName = annotation.String
				break
			}
		}

		enumInfo.Members[member.Name] = memberInfo
	}

	p.Enums[fullName] = enumInfo
	p.Enums[enumType.Name] = enumInfo // Also store by short name
}

// parseEntityType processes an entity type definition
func (p *MetadataParser) parseEntityType(entityType EntityType, namespace string) {
	entityInfo := &EntityInfo{
		Name:       entityType.Name,
		Properties: make(map[string]*PropertyInfo),
		BaseType:   entityType.BaseType,
		IsBaseType: entityType.BaseType != "",
	}

	// Process properties
	for _, property := range entityType.Properties {
		propInfo := &PropertyInfo{
			Name:         property.Name,
			Type:         property.Type,
			IsRequired:   property.Nullable == "false",
			IsCollection: strings.HasPrefix(property.Type, "Collection("),
		}

		// Determine if this is an enum type
		if enumType := p.extractEnumType(property.Type); enumType != "" {
			propInfo.EnumType = enumType
		}

		entityInfo.Properties[property.Name] = propInfo
	}

	p.Entities[entityType.Name] = entityInfo
}

// extractEnumType extracts enum type name from a property type
func (p *MetadataParser) extractEnumType(propType string) string {
	// Handle Collection(EnumType) format
	if strings.HasPrefix(propType, "Collection(") && strings.HasSuffix(propType, ")") {
		inner := propType[11 : len(propType)-1] // Remove "Collection(" and ")"
		if strings.Contains(inner, "org.reso.metadata.enums.") {
			return strings.TrimPrefix(inner, "org.reso.metadata.enums.")
		}
		return inner
	}

	// Handle direct enum type
	if strings.Contains(propType, "org.reso.metadata.enums.") {
		return strings.TrimPrefix(propType, "org.reso.metadata.enums.")
	}

	return ""
}

// GetEntityInfo returns information about a specific entity
func (p *MetadataParser) GetEntityInfo(entityName string) (*EntityInfo, bool) {
	entity, exists := p.Entities[entityName]
	return entity, exists
}

// GetEnumInfo returns information about a specific enum
func (p *MetadataParser) GetEnumInfo(enumName string) (*EnumInfo, bool) {
	enum, exists := p.Enums[enumName]
	return enum, exists
}

// GetEntityNames returns sorted list of all entity names
func (p *MetadataParser) GetEntityNames() []string {
	var names []string
	for name := range p.Entities {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// GetEnumNames returns sorted list of all enum names
func (p *MetadataParser) GetEnumNames() []string {
	var names []string
	seen := make(map[string]bool)

	for name := range p.Enums {
		if !seen[name] {
			names = append(names, name)
			seen[name] = true
		}
	}
	sort.Strings(names)
	return names
}

// GetFieldsByCategory returns fields organized by logical categories
func (p *MetadataParser) GetFieldsByCategory(entityName string) map[string][]string {
	entity, exists := p.Entities[entityName]
	if !exists {
		return nil
	}

	categories := make(map[string][]string)

	for fieldName := range entity.Properties {
		category := p.categorizeField(fieldName)
		categories[category] = append(categories[category], fieldName)
	}

	// Sort fields within each category
	for category := range categories {
		sort.Strings(categories[category])
	}

	return categories
}

// categorizeField categorizes a field based on its name patterns
func (p *MetadataParser) categorizeField(fieldName string) string {
	lower := strings.ToLower(fieldName)

	// Identification fields
	if strings.Contains(lower, "key") || strings.Contains(lower, "id") ||
		strings.Contains(lower, "mlsid") || fieldName == "ListingKey" || fieldName == "ListingId" {
		return "Identification"
	}

	// Address and location fields
	if strings.Contains(lower, "street") || strings.Contains(lower, "city") ||
		strings.Contains(lower, "state") || strings.Contains(lower, "postal") ||
		strings.Contains(lower, "address") || strings.Contains(lower, "latitude") ||
		strings.Contains(lower, "longitude") || strings.Contains(lower, "area") {
		return "Address & Location"
	}

	// Pricing fields
	if strings.Contains(lower, "price") || strings.Contains(lower, "tax") ||
		strings.Contains(lower, "cost") || strings.Contains(lower, "expense") ||
		strings.Contains(lower, "fee") || strings.Contains(lower, "income") {
		return "Pricing & Financial"
	}

	// Property characteristics
	if strings.Contains(lower, "bedroom") || strings.Contains(lower, "bathroom") ||
		strings.Contains(lower, "room") || strings.Contains(lower, "area") ||
		strings.Contains(lower, "year") || strings.Contains(lower, "built") ||
		strings.Contains(lower, "stories") || strings.Contains(lower, "lot") {
		return "Property Details"
	}

	// Agent/Office information
	if strings.Contains(lower, "agent") || strings.Contains(lower, "office") ||
		strings.Contains(lower, "member") || strings.Contains(lower, "broker") {
		return "Agent & Office Info"
	}

	// Status and dates
	if strings.Contains(lower, "status") || strings.Contains(lower, "timestamp") ||
		strings.Contains(lower, "date") || strings.Contains(lower, "time") ||
		strings.Contains(lower, "market") || strings.Contains(lower, "modification") {
		return "Status & Dates"
	}

	// Features and amenities
	if strings.Contains(lower, "feature") || strings.Contains(lower, "amenity") ||
		strings.Contains(lower, "appliance") || strings.Contains(lower, "heating") ||
		strings.Contains(lower, "cooling") || strings.Contains(lower, "parking") ||
		strings.Contains(lower, "pool") || strings.Contains(lower, "garage") ||
		strings.Contains(lower, "fireplace") || strings.HasSuffix(lower, "yn") {
		return "Features & Amenities"
	}

	// Media related
	if strings.Contains(lower, "media") || strings.Contains(lower, "photo") ||
		strings.Contains(lower, "video") || strings.Contains(lower, "image") ||
		strings.Contains(lower, "virtual") || strings.Contains(lower, "url") {
		return "Media & Marketing"
	}

	// Default category
	return "Other"
}

// GenerateEntityGuide generates dynamic entity documentation
func (p *MetadataParser) GenerateEntityGuide() string {
	var guide strings.Builder
	guide.WriteString("# RESO Entities Guide (Generated from Metadata)\n\n")

	entityNames := p.GetEntityNames()
	for _, entityName := range entityNames {
		entity := p.Entities[entityName]

		guide.WriteString(fmt.Sprintf("## %s Entity\n", entityName))

		if entity.IsBaseType {
			guide.WriteString(fmt.Sprintf("**Base Type**: %s\n", entity.BaseType))
		}

		guide.WriteString(fmt.Sprintf("**Total Fields**: %d\n\n", len(entity.Properties)))

		// Show key fields
		keyFields := p.getKeyFields(entity)
		if len(keyFields) > 0 {
			guide.WriteString("**Key Fields**: ")
			guide.WriteString(strings.Join(keyFields, ", "))
			guide.WriteString("\n\n")
		}

		// Show field categories
		categories := p.GetFieldsByCategory(entityName)
		for category, fields := range categories {
			if len(fields) > 0 {
				guide.WriteString(fmt.Sprintf("**%s**: ", category))
				if len(fields) > 10 {
					guide.WriteString(fmt.Sprintf("%s... (%d total fields)\n",
						strings.Join(fields[:10], ", "), len(fields)))
				} else {
					guide.WriteString(fmt.Sprintf("%s\n", strings.Join(fields, ", ")))
				}
			}
		}
		guide.WriteString("\n")
	}

	return guide.String()
}

// GenerateFieldsGuide generates dynamic fields documentation
func (p *MetadataParser) GenerateFieldsGuide(entityName string) string {
	entity, exists := p.Entities[entityName]
	if !exists {
		return fmt.Sprintf("Entity '%s' not found in metadata", entityName)
	}

	var guide strings.Builder
	guide.WriteString(fmt.Sprintf("# %s Entity Fields (Generated from Metadata)\n\n", entityName))

	categories := p.GetFieldsByCategory(entityName)

	for category, fields := range categories {
		if len(fields) == 0 {
			continue
		}

		guide.WriteString(fmt.Sprintf("## %s\n\n", category))

		for _, fieldName := range fields {
			prop := entity.Properties[fieldName]
			guide.WriteString(fmt.Sprintf("- **%s** (%s)", fieldName, p.formatType(prop.Type)))

			if prop.IsRequired {
				guide.WriteString(" *Required*")
			}

			if prop.EnumType != "" {
				guide.WriteString(fmt.Sprintf(" - Enum: %s", prop.EnumType))
			}

			guide.WriteString("\n")
		}
		guide.WriteString("\n")
	}

	return guide.String()
}

// GenerateEnumsGuide generates dynamic enums documentation
func (p *MetadataParser) GenerateEnumsGuide() string {
	var guide strings.Builder
	guide.WriteString("# RESO Enum Values (Generated from Metadata)\n\n")

	enumNames := p.GetEnumNames()

	// Focus on most commonly used enums first
	priorityEnums := []string{
		"StandardStatus", "PropertyType", "PropertySubType",
		"MediaCategory", "StateOrProvince", "AreaUnits",
	}

	// Add priority enums first
	for _, enumName := range priorityEnums {
		if enumInfo, exists := p.Enums[enumName]; exists {
			guide.WriteString(p.formatEnumSection(enumInfo))
		}
	}

	// Add remaining enums
	for _, enumName := range enumNames {
		// Skip if already added as priority
		isPriority := false
		for _, priority := range priorityEnums {
			if enumName == priority {
				isPriority = true
				break
			}
		}
		if isPriority {
			continue
		}

		if enumInfo, exists := p.Enums[enumName]; exists {
			guide.WriteString(p.formatEnumSection(enumInfo))
		}
	}

	return guide.String()
}

// formatEnumSection formats an enum for documentation
func (p *MetadataParser) formatEnumSection(enumInfo *EnumInfo) string {
	var section strings.Builder
	section.WriteString(fmt.Sprintf("## %s\n\n", enumInfo.Name))

	// Get sorted member names
	var memberNames []string
	for memberName := range enumInfo.Members {
		memberNames = append(memberNames, memberName)
	}
	sort.Strings(memberNames)

	for _, memberName := range memberNames {
		member := enumInfo.Members[memberName]
		section.WriteString(fmt.Sprintf("- **%s**", member.Name))

		if member.StandardName != "" && member.StandardName != member.Name {
			section.WriteString(fmt.Sprintf(" (%s)", member.StandardName))
		}

		if member.Value != "" {
			section.WriteString(fmt.Sprintf(" - Value: %s", member.Value))
		}

		section.WriteString("\n")
	}
	section.WriteString("\n")

	return section.String()
}

// getKeyFields extracts key field names from entity
func (p *MetadataParser) getKeyFields(entity *EntityInfo) []string {
	var keyFields []string

	// Look for common key patterns
	for fieldName := range entity.Properties {
		if strings.Contains(strings.ToLower(fieldName), "key") ||
			strings.Contains(strings.ToLower(fieldName), "id") {
			keyFields = append(keyFields, fieldName)
		}
	}

	sort.Strings(keyFields)
	return keyFields
}

// formatType formats a property type for display
func (p *MetadataParser) formatType(propType string) string {
	// Clean up common type patterns
	if strings.HasPrefix(propType, "Collection(") {
		inner := propType[11 : len(propType)-1]
		return fmt.Sprintf("Collection of %s", p.formatType(inner))
	}

	if strings.HasPrefix(propType, "Edm.") {
		return strings.TrimPrefix(propType, "Edm.")
	}

	if strings.Contains(propType, "org.reso.metadata.enums.") {
		return strings.TrimPrefix(propType, "org.reso.metadata.enums.")
	}

	return propType
}

// GetCommonFields returns most commonly used fields for an entity
func (p *MetadataParser) GetCommonFields(entityName string) []string {
	entity, exists := p.Entities[entityName]
	if !exists {
		return nil
	}

	// Define common field patterns based on entity type
	switch entityName {
	case "Property":
		return p.getCommonPropertyFields(entity)
	case "Member":
		return p.getCommonMemberFields(entity)
	case "Office":
		return p.getCommonOfficeFields(entity)
	case "Media":
		return p.getCommonMediaFields(entity)
	default:
		// Return first 20 fields as common fields
		var fields []string
		for fieldName := range entity.Properties {
			fields = append(fields, fieldName)
			if len(fields) >= 20 {
				break
			}
		}
		sort.Strings(fields)
		return fields
	}
}

// getCommonPropertyFields returns commonly used Property fields
func (p *MetadataParser) getCommonPropertyFields(entity *EntityInfo) []string {
	commonPatterns := []string{
		"ListingKey", "ListingId", "StandardStatus", "MlsStatus",
		"ListPrice", "ClosePrice", "OriginalListPrice",
		"StreetNumber", "StreetName", "City", "StateOrProvince", "PostalCode", "UnparsedAddress",
		"BedroomsTotal", "BathroomsTotal", "LivingArea", "YearBuilt", "LotSizeSquareFeet",
		"PropertyType", "PropertySubType",
		"ListAgentFullName", "ListAgentEmail", "ListAgentMlsId", "ListOfficeName",
		"OnMarketTimestamp", "ModificationTimestamp", "DaysOnMarket",
		"PublicRemarks", "PhotosCount", "Latitude", "Longitude",
	}

	var existing []string
	for _, field := range commonPatterns {
		if _, exists := entity.Properties[field]; exists {
			existing = append(existing, field)
		}
	}

	return existing
}

// getCommonMemberFields returns commonly used Member fields
func (p *MetadataParser) getCommonMemberFields(entity *EntityInfo) []string {
	commonPatterns := []string{
		"MemberKey", "MemberMlsId", "MemberFullName", "MemberFirstName", "MemberLastName",
		"MemberEmail", "MemberDirectPhone", "MemberMobilePhone",
		"OfficeKey", "OfficeMlsId", "OfficeName",
		"MemberDesignation", "MemberStatus", "ModificationTimestamp",
	}

	var existing []string
	for _, field := range commonPatterns {
		if _, exists := entity.Properties[field]; exists {
			existing = append(existing, field)
		}
	}

	return existing
}

// getCommonOfficeFields returns commonly used Office fields
func (p *MetadataParser) getCommonOfficeFields(entity *EntityInfo) []string {
	commonPatterns := []string{
		"OfficeKey", "OfficeMlsId", "OfficeName",
		"OfficePhone", "OfficeEmail", "OfficeFax",
		"OfficeAddress1", "OfficeAddress2", "OfficeCity", "OfficeStateOrProvince", "OfficePostalCode",
		"ModificationTimestamp",
	}

	var existing []string
	for _, field := range commonPatterns {
		if _, exists := entity.Properties[field]; exists {
			existing = append(existing, field)
		}
	}

	return existing
}

// getCommonMediaFields returns commonly used Media fields
func (p *MetadataParser) getCommonMediaFields(entity *EntityInfo) []string {
	commonPatterns := []string{
		"MediaKey", "ResourceRecordKey", "ResourceRecordID",
		"MediaURL", "MediaType", "MediaCategory", "MediaStatus",
		"Permission", "Order", "ModificationTimestamp",
		"ShortDescription", "LongDescription",
	}

	var existing []string
	for _, field := range commonPatterns {
		if _, exists := entity.Properties[field]; exists {
			existing = append(existing, field)
		}
	}

	return existing
}
