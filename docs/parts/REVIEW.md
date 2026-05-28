# parts Package — Feature Overview

## 1. Slide Module (slide.go, slide_types.go)

**Core responsibility**: Generating and parsing slide XML structures

| Type / Function | Description |
|----------|------|
| SlidePart | Slide part, corresponds to /ppt/slides/slideN.xml |
| SlideLayoutPart | Layout part, corresponds to /ppt/slideLayouts/slideLayoutN.xml |
| ShapeIDAllocator | Shape ID allocator (single-threaded) |
| ShapeIDAllocatorSync | Shape ID allocator (thread-safe) |
| XMLWriter / XMLWriterPool | Streaming XML write helpers |
| XML struct types: XSlide, XSpTree, XSp, XPicture, XGraphicFrame, XTable, XTextBody, etc. |

**Note:** Relationship management in `SlidePart` uses `opc.Relationships`; deduplication logic is encapsulated through the `AddImageRel`/`AddMediaRel`/`AddChartRel`/`AddTableRel` methods.

## 2. Master Module (master.go, master_types.go, master_parser.go, master_cache.go)
**Core responsibility:** Read-only data structures, parsing, and caching for slide masters and layouts

| Type / Function | Description |
|----------|------|
| MasterManager | Master manager (facade pattern) |
| MasterCache | Master/layout cache (concurrency-safe reads) |
| SlideMasterData | Read-only master data |
| SlideLayoutData | Read-only layout data |
| Placeholder | Placeholder definition |
| Background | Background definition |
| ParseLayout() | Parse layout XML |
| ParseMaster() | Parse master XML |
| Enum types: PlaceholderType, BackgroundType, SlideLayoutType | |

## 3. Presentation Module (presentation.go)
**Core responsibility**: Presentation root node

| Type / Function | Description |
| ---------- | ------ |
| PresentationPart | Presentation part, corresponds to /ppt/presentation.xml |
| SlideSize | Slide size (EMU units) |
| StandardSlideSizes | Standard sizes (16:9, 4:3) |
| EMUFromPoints() etc. | EMU unit conversion functions |
| XML struct types: XPresentation, XSldIdLst, XSldMasterIdLst | |

## 4. Media Module (media.go, media_manager.go)
**Core responsibility**: Media resource management

| Type / Function | Description |
| ---------- | ------ |
| MediaResource | Media resource (image/audio/video) |
| MediaManager | Media resource manager (concurrency-safe cache) |
| MediaType | Media type enum |
| NewMediaResourceFromBytes() | Create resource from bytes |
| NewMediaResourceFromReader() | Create resource from Reader |
| Features: deduplication, multi-index (rID/fileName/target/hash), MIME type inference |

## 5. Relationship Module (relationship.go)
**Core responsibility**: OPC relationship XML structure definitions (pure DTO)

| Type / Function | Description |
| ---------- | ------ |
| XMLRelationships | Relationship collection (for serializing/deserializing .rels files) |
| XMLRelationship | Single relationship |
| ParseRelationships() | Parse relationship XML |
| Constants: RelTypeImage, RelTypeSlide, RelTypeSlideLayout, etc. |

**Note:** Relationship management logic has been moved to `opc.Relationships` in the `opc` layer; this module retains only the XML DTO for reading and writing .rels files.

## 6. Theme Module (theme.go, theme_types.go, theme_default.go)
**Core responsibility**: Theme template management (minimally processed)

| Type / Function | Description |
| ---------- | ------ |
| ThemePart | Theme part, corresponds to /ppt/theme/themeN.xml |
| XTheme, XThemeElements | Theme XML structures |
| XColorScheme, XColorVariant | Color scheme |
| XFontScheme, XFontCollection | Font scheme |
| XFmtScheme | Format scheme (raw data preserved via InnerXML) |
| DefaultThemeXML | Complete Office theme template constant |
| DefaultTheme() | Get the default theme (lazy-loaded singleton) |
| CloneTheme() | Clone a theme (deep copy) |
| GetThemeColor/GetThemeColorRGB/GetThemeColorType | Color accessor methods |
| SetThemeColorRGB/SetThemeColorSystem | Color setter methods |
| SetThemeMajorFont/SetThemeMinorFont/SetThemeScriptFont | Font setter methods |

**Design principles:**
1. **Entry point, not the primary concern**: Color/font read-write methods are provided, but themes are highly complex — deep customisation is not recommended.
2. **Template-first**: `DefaultThemeXML` (a complete Office theme) + `CloneTheme()` ensure the generated PPTX structure is complete.
3. **Data preservation**: FmtScheme uses `InnerXML` to retain the original XML, avoiding data loss during parsing.

## 7. AppProps Module (appprops.go, appprops_types.go)
**Core responsibility**: Application properties (company, manager, etc.)

| Type / Function | Description |
| ---------- | ------ |
| AppPropsPart | Application properties part, corresponds to /docProps/app.xml |
| XMLAppProps | Application properties XML structure |
| GetAppCompany/SetAppCompany | Company name read/write |
| GetAppManager/SetAppManager | Manager read/write |
| GetAppSlideCount/SetAppSlideCount | Slide count read/write |
| SetAppWordCount/SetAppTotalTime | Word count / editing time setters |
| HeadingPairs/TitlesOfParts | Heading pairs / part titles (preserved via InnerXML) |

**Design notes:**
- The OOXML spec requires company, manager, and similar metadata to be written to `/docProps/app.xml`.
- Methods uniformly use the `App` prefix to avoid conflicts with Go keywords.
- HeadingPairs and TitlesOfParts use InnerXML to preserve the original structure and avoid complex parsing.

## 8. CoreProps Module (coreprops.go)
**Core responsibility**: Core properties

| Type / Function | Description |
| ---------- | ------ |
| XMLCoreProperties | Core properties structure |
| XMLW3CDTFDate | W3CDTF date format |
| ParseCoreProperties() | Parse core properties |

## 9. Chart Module (chart.go, chart_types.go)
**Core responsibility**: Chart parts (template + placeholder strategy)

| Type / Function | Description |
| ---------- | ------ |
| ChartPart | Chart part, corresponds to /ppt/charts/chartN.xml |
| ChartType | Chart type enum (Bar/Pie/Line/Area/Scatter/Doughnut) |
| ChartTemplateBar/Pie/Line... | Pre-defined chart template constants |
| SetTemplate/SetRawXML | Set chart template / raw XML |
| ReplacePlaceholder | Replace a single placeholder |
| ReplacePlaceholders | Batch-replace placeholders |
| SetExternalDataRID/GetExternalDataRID | External Excel data reference |
| HasExternalData | Check whether an external data reference exists |

**Design strategy:**
- **Template + placeholders**: Avoids mapping complex chart XML (hundreds of element combinations) to Go structs.
- **Pre-defined templates**: Provides common chart types (bar, pie, line, etc.).
- **Placeholder replacement**: `{{CHART_TITLE}}`, `{{CATEGORIES}}`, `{{SERIES_VALUES}}`, etc.
- **Two approaches:**
  - Route C (no Excel): Data is embedded directly in `strCache`/`numCache` with no external dependency.
  - Route A/B (with Excel): References an embedded Excel file via `externalData`.

**Common placeholders:**
| Placeholder | Description |
|--------|------|
| `{{CHART_TITLE}}` | Chart title |
| `{{SERIES_NAME}}` | Series name |
| `{{CATEGORIES}}` | Category label XML fragment |
| `{{SERIES_VALUES}}` | Value XML fragment |
| `{{CAT_COUNT}}` | Category count |
| `{{CAT_COUNT_PLUS_1}}` | Category count + 1 (used in Excel formulas) |

## 10. Embedding Module (embedding.go)
**Core responsibility**: Embedded data parts

| Type / Function | Description |
| ---------- | ------ |
| EmbeddingPart | Embedding part, corresponds to /ppt/embeddings/*.xlsx |
| EmbeddingType | Embedding type enum (Excel/Word/Other) |
| Data/SetData | Binary data read/write |
| SetDataReader | Set data from Reader |
| DetectEmbeddingType | Detect type from file name |

**Design notes:**
- Embedded data is a binary file (e.g. Excel); no XML parsing is performed.
- Reader/Writer interfaces are provided for streaming use.

## 11. XML Utilities Module (xmlutils.go)
**Core responsibility**: XML processing utilities

| Function / Constant | Description |
|----------|------|
| XMLDeclaration | XML declaration header constant |
| StripNamespacePrefixes() | Remove namespace prefixes |

## 12. XML Master Models (xml_master_models.go)
**Core responsibility**: Intermediate structures used when parsing master/layout XML

| Type | Description |
|------|------|
| XMLOffset, XMLExtents, XMLTransform | Position / size / transform |
| XMLPlaceholder | Placeholder |
| XMLShape, XMLShapeTree | Shape / shape tree |
| XMLBackground, XMLFillProperties | Background / fill |
| XMLSlideLayout, XMLSlideMaster | Layout / master |

---

# Architecture Layering Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                        opc package                          │
│  Responsibility: generic OPC spec implementation            │
│                  (package, parts, relationship management)  │
│  - PackURI: path handling                                   │
│  - Relationships: thread-safe relationship management +     │
│                   atomic ID allocation                      │
│  - Part/Package: foundational structures for parts/packages │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                        parts package                        │
│  Responsibility: PPTX-specific XML structure definitions +  │
│                  serialization / deserialization            │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │   Slide      │  │   Master     │  │ Presentation │      │
│  │  Slide XML   │  │ Master/Layout│  │ Presentation │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│                                                             │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │    Theme     │  │    Media     │  │ Relationship │      │
│  │  Theme tmpl  │  │  Media res.  │  │   XML DTO    │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│                                                             │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │  CoreProps   │  │     App      │  │   XMLUtils   │      │
│  │ Core props   │  │  App props   │  │  XML utils   │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│                                                             │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                       slide package                         │
│  Responsibility: high-level business logic                  │
│                  (SlideBuilder, MediaManager)               │
└─────────────────────────────────────────────────────────────┘
```

---

# File Inventory

| File | Lines | Main Contents |
|------|------|----------|
| slide.go | ~1700 | SlidePart, XMLWriter, WriteXML methods |
| slide_types.go | ~450 | XML struct type definitions |
| theme.go | ~490 | ThemePart, theme read/write methods |
| theme_types.go | ~180 | Theme XML struct types |
| theme_default.go | ~260 | Default theme template, CloneTheme |
| appprops.go | ~270 | AppPropsPart, application property read/write methods |
| appprops_types.go | ~100 | Application properties XML struct types |
| chart.go | ~130 | ChartPart, chart read/write methods |
| chart_types.go | ~120 | Chart XML struct types |
| embedding.go | ~130 | EmbeddingPart, embedded data read/write |
| media_manager.go | 460 | MediaManager |
| presentation.go | 393 | PresentationPart |
| master_parser.go | 344 | Master/layout parser |
| master_types.go | 358 | Master data structures |
| master_cache.go | 275 | MasterCache |
| xml_master_models.go | 272 | Intermediate XML structures for masters |
| master.go | 255 | MasterManager |
| media.go | 244 | MediaResource |
| relationship.go | 179 | XMLRelationships (pure DTO) |
| coreprops.go | 161 | XMLCoreProperties |
| xmlutils.go | 89 | XML utility functions |
