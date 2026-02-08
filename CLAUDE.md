# CLAUDE.md - dicomforge

Go CLI generating valid DICOM series for testing medical imaging platforms. Modalities: MR, CT, CR, DX, US, MG.

## Build & test

```bash
go build -o dicomforge ./cmd/dicomforge/                              # build
go test $(go list ./... | grep -v /tests/e2e) -v -race -short         # CI tests
go test ./tests/e2e/... -v                                            # E2E (needs dcmtk)
go vet ./... && staticcheck ./... && golangci-lint run                 # lint
gosec -exclude=G115,G301,G302,G304,G404 ./...                         # security
```

Version: `-ldflags "-X main.version=X.Y.Z"`. CI: Go 1.22/1.23/1.24. Nix flake for dev shell.

## Before pushing to a PR branch

Always run lint, security and tests locally before pushing:

```bash
go vet ./... && staticcheck ./... && golangci-lint run                 # lint
gosec -exclude=G115,G301,G302,G304,G404 ./...                         # security
go test $(go list ./... | grep -v /tests/e2e) -v -race -short         # tests
```

## Project layout

```
cmd/dicomforge/main.go        CLI flags → GeneratorOptions → GenerateDICOMSeries() → OrganizeFilesIntoDICOMDIR()
cmd/dicomforge/wizard/         TUI (Bubbletea/huh). config.go=YAML, convert.go=state→opts, types/=WizardState
internal/dicom/generator.go    Core: GeneratorOptions, imageTask, GeneratedFile, worker pool
internal/dicom/metadata.go     mustNewElement(), GenerateMetadata()
internal/dicom/dicomdir.go     OrganizeFilesIntoDICOMDIR(), PT/ST/SE hierarchy, DICOMDIR offset patching
internal/dicom/corruption/     types.go(Config,ParseTypes) applicator.go siemens.go ge.go philips.go malformed.go
internal/dicom/edgecases/      types.go(Config,ParseTypes) applicator.go specialchars.go longnames.go dates.go variedids.go missingtags.go
internal/dicom/modalities/     modality.go(Generator interface) mr.go ct.go cr.go dx.go us.go mg.go helpers.go series_templates.go
internal/image/                pixel.go(GenerateSingleImage) overlay.go(AddTextOverlay, 8/16-bit)
internal/util/                 uid.go names.go size.go clinical.go institutions.go priority.go series_range.go tagparser.go tagregistry.go
tests/                         integration, reproducibility, validation, compatibility, performance, errors, e2e/
```

## Key types & interfaces

**GeneratorOptions** (generator.go): NumImages, TotalSize, OutputDir, Seed, NumStudies, NumPatients, Workers, Modality, SeriesPerStudy(util.SeriesRange), StudyDescriptions, Institution, Department, BodyPart, Priority(util.Priority), VariedMetadata, CustomTags(util.ParsedTags), EdgeCaseConfig(edgecases.Config), CorruptionConfig(corruption.Config), Quiet, ProgressCallback, PredefinedPatients([]PredefinedPatient)

**PredefinedPatient/Study/Series**: Fully pre-configured patient hierarchy from wizard YAML. Patient{Name,ID,BirthDate,Sex,Studies}, Study{Description,Date,AccessionNumber,Institution,Department,BodyPart,Priority,ReferringPhysician,Series}, Series{Description,Protocol,Orientation,ImageCount}

**modalities.Generator interface**: Modality(), SOPClassUID(), Scanners(), GenerateSeriesParams(Scanner,*rand.Rand)→SeriesParams, PixelConfig(), AppendModalityElements(*dicom.Dataset,SeriesParams), WindowPresets()

**SeriesParams**: Common(WindowCenter/Width,PixelSpacing,SliceThickness) + MR(EchoTime,RepetitionTime,FlipAngle,SequenceName,MagneticFieldStrength,ImagingFrequency) + CT(KVP,XRayTubeCurrent,ConvolutionKernel,RescaleIntercept/Slope,GantryTilt) + CR/DX(ViewPosition,ImagerPixelSpacing,DistanceSourceToDetector/Patient,Exposure,ExposureTime) + US(TransducerType,TransducerFrequency) + MG(ImageLaterality,AnodeTargetMaterial,FilterMaterial,CompressionForce,OrganDose)

**PixelConfig**: BitsAllocated/Stored/HighBit/PixelRepresentation(uint16), MinValue/MaxValue/BaseValue(int). MR=12bit(0-4095), CT=16bit signed(-1024 to 3071), CR=12bit, DX=14bit, US=8bit(0-255), MG=14bit

**GeneratedFile**: Path, StudyUID, SeriesUID, SOPInstanceUID, PatientID, StudyID, SeriesNumber, InstanceNumber, InstanceInStudy

## Generation pipeline

1. Parse & validate options, ParseSize() → bytes, CalculateDimensions() → width/height
2. Seed: explicit or FNV64a hash of OutputDir name
3. Create edgecases.Applicator + corruption.Applicator if enabled
4. Generate/load patient data (PredefinedPatients or auto-generated)
5. **Phase 1 (sequential)**: Build []imageTask — for each study: deterministic UIDs via GenerateDeterministicUID(OutputDir+index), scanner selection, series params, metadata elements, corruption elements, pixel seed
6. **Phase 2 (parallel)**: Worker pool (goroutines, taskChan/resultChan, default=NumCPU). Each worker: RNG from pixelSeed → pixel generation → text overlay → DICOM write → optional malformed patching
7. OrganizeFilesIntoDICOMDIR: group by PatientID→StudyUID→SeriesUID, rename to PT%06d/ST%06d/SE%06d/IM%06d, create DICOMDIR with binary offset patching

## Features inventory

**6 modalities** with realistic scanners:
- MR: Siemens(Avanto 1.5T,Skyra 3.0T), GE(Signa,Discovery), Philips(Achieva,Ingenia). Sequences: T1_MPRAGE,T1_SE,T2_FSE,T2_FLAIR. Series templates per body part (brain,knee,spine,abdomen)
- CT: Siemens,GE,Philips,Canon (16-320 detector rows). KVP 80-140. Kernels: SOFT,STANDARD,BONE,LUNG. Window presets: BRAIN,SUBDURAL,BONE,LUNG,MEDIASTINUM,ABDOMEN,LIVER
- CR: Fujifilm,Carestream,Agfa,Konica,Philips. Views: AP,PA,LAT,LL,RL
- DX: Siemens,GE,Philips,Carestream,Canon,Fujifilm. Finer pixel spacing than CR
- US: GE,Philips,Siemens,Canon,Samsung,Hitachi. Transducers: LINEAR(7-15MHz),CONVEX(2-6MHz),PHASED(2-5MHz). Only 8-bit modality
- MG: Hologic,GE,Siemens,Fujifilm,Philips,IMS Giotto. Views: CC,MLO,ML,LM. Laterality L/R. Anode: Mo,Rh,W. 14-bit, KVP 25-34

**4 corruption types** (--corrupt):
- siemens-csa: Private creator (0029,0010)="SIEMENS CSA HEADER", CSA Image(0029,1010) + Series(0029,1020) headers with "SV10" binary format, crash-trigger SQ(0029,1102)
- ge-private: Creators GEMS_IDEN_01(0009,0010) + GEMS_PARM_01(0043,0010), software version(0009,10E3), multi-valued diffusion params(0043,1039)
- philips-private: Creators "Philips Imaging DD 001"(2001,0010) + "Philips MR Imaging DD 001/005"(2005,0010/0011), nested SQ(2005,100E) with scale slope/intercept
- malformed-lengths: Placeholder (0071,0010)→patched to (0070,0253) FL with length not multiple of 4, PixelData(7FE0,0010) OW with odd byte count. Post-processed via PatchMalformedLengths() binary file rewrite

**5 edge case types** (--edge-cases N --edge-case-types):
- special-chars: Names with accents, hyphens, apostrophes (Jean-Pierre, Müller-Schmidt, O'Connor, François, etc.)
- long-names: 64-char DICOM limit patient names and IDs
- missing-tags: Omit 1-3 random optional DICOM tags
- old-dates: Birth dates 1900-1950, partial dates (YYYYMM format), future study dates (25% chance)
- varied-ids: Patient IDs with dashes, letters, spaces, max length

**YAML config**: Load(--config)/Save(--save-config). Structure: global{modality,total_images,total_size,output,seed,num_patients,studies_per_patient,series_per_study} + patients[]{name,id,birth_date,sex,studies[]{description,date,accession,institution,department,body_part,priority,referring_physician,custom_tags,series[]{description,protocol,orientation,images,custom_tags}}}

**Custom tags** (--tag "Name=Value"): 25 supported tags across 4 scopes (patient/study/series/image). tagregistry.go has fuzzy matching with Levenshtein distance suggestions

**Patient names**: 80% English / 20% French. 400+ names in pools. Format "LASTNAME^FIRSTNAME". Physician: 50% with "Dr" prefix

**Clinical context**: French clinical indications per body part. Body parts per modality. Protocol names per modality+body part. 15 hospitals (FR+US), 10 departments

**Deterministic UIDs**: SHA256(seed string) → DICOM UID prefix "1.2.826.0.1.3680043.8.498" + numeric segments, max 64 chars

**Pixel generation**: Radial gradient + multi-scale noise (large/medium/fine), deterministic from pixelSeed. Text overlay "File X/Y" scaled to 30% image width, black outline + white text. 8-bit path (US) and 16-bit path (all others)

## Code conventions

- `gofmt` + `golangci-lint`. Explicit error returns (panic only in `mustNewElement()`)
- Test: `TestFunctionName_Scenario(t *testing.T)`, `t.TempDir()` for isolation
- DICOM tags via `suyashkumar/dicom` (`tag.PatientName`, etc.)
- Private elements via `mustNewPrivateElement(tag, rawVR, data)` for explicit VR control
- gosec exclusions: G115(bounded ints), G301/G302(standard perms), G304(user paths), G404(reproducibility RNG)

## Dependencies

`suyashkumar/dicom` v1.1.0 (DICOM I/O), `golang.org/x/image` v0.34.0 (overlay), `cucumber/godog` v0.15.1 (BDD), `charmbracelet/bubbletea`+`huh`+`lipgloss` (TUI)

## CLI flags

Required: `--num-images N --total-size SIZE`
Optional: `--output --seed --modality --num-studies --num-patients --series-per-study --workers --institution --department --body-part --priority --varied-metadata --tag --edge-cases --edge-case-types --corrupt --config --save-config --interactive/-i --version --help`
Subcommand: `wizard [--from config.yaml]`
