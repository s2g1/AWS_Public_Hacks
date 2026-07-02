# TECHNICAL AND COST PROPOSAL
## S2G Technologies, Inc.
### In Response To: Solicitation No. W15QKN-25-R-0042
### Autonomous Drone Swarm Battlefield Mapping System

**PROPRIETARY & CONFIDENTIAL**

---

## Proposal Summary

| Field | Value |
|-------|-------|
| **Offeror** | S2G Technologies, Inc. |
| **CAGE Code** | 8X4T2 |
| **Solicitation** | W15QKN-25-R-0042 |
| **Proposal Date** | August 1, 2025 |
| **Validity Period** | 180 Days from Submission |
| **Contract Type** | Cost-Plus-Fixed-Fee (CPFF) |
| **Total Proposed Cost (Base)** | $2,847,400 |

---

## VOLUME I – EXECUTIVE SUMMARY

S2G Technologies proposes **Project ARGUS** (Autonomous Reconnaissance Grid – Unified Swarm), employing a three-tier AWS cloud and edge architecture to achieve real-time, collaborative battlefield picture generation across 10 km² while operating under EMCON constraints.

### Key Discriminators
1. AWS Greengrass Edge Runtime on each swarm agent for onboard AI inference
2. Pre-mission AWS Ground Station integration for threat environment download
3. Amazon Rekognition Custom Labels + fine-tuned vision transformers for edge terrain extraction
4. AWS IoT Core mesh backbone for post-EMCON data consolidation
5. ATAK plugin and NITF/GeoTIFF export native to GCS delivery

### Mission Architecture
- **Pre-Mission Phase**: AWS Ground Station + S3 for mission package staging
- **EMCON Execution Phase**: Onboard Greengrass inference, optical inter-agent mesh, zero cloud dependency
- **Consolidation Phase**: AWS IoT Core + AWS Batch photogrammetry pipeline

---

## VOLUME II – TECHNICAL APPROACH

### 2.1 System Architecture – AWS-Native Stack

#### Cloud Control Tier (Pre/Post Mission)
- AWS Ground Station: Satellite link for threat data
- Amazon S3: Mission packages (waypoints, terrain, ML weights, ROE)
- AWS Lambda + Step Functions: Automated mission planning
- Amazon SageMaker: RL model training, CV fine-tuning
- AWS Batch: Post-mission photogrammetry (OpenDroneMap)
- Amazon CloudWatch: Telemetry aggregation

#### Edge Swarm Tier (In-Mission, EMCON Compliant)
- AWS IoT Greengrass Core (per agent): Local ML inference, inter-agent routing
- Amazon Rekognition Custom Labels (edge): Real-time terrain classification at 10 Hz
- AWS Panorama: Multi-sensor data fusion
- Custom FSOC protocol: Free Space Optical Communication (850nm, zero RF)
- Visual-Inertial Odometry: GPS-denied navigation via Kalman filter fusion

#### GCS Integration Tier
- AWS IoT Core: Bidirectional telemetry (MQTT over LPI/LPD)
- Amazon Kinesis Video Streams: Multi-platform video downlink
- Custom ATAK Plugin: Real-time swarm overlay
- Export module: NITF/GeoTIFF/KMZ on mission consolidation

### 2.2 EMCON Compliance Strategy

Full EMCON Level II via pre-mission staging. Inter-agent coordination through proprietary FSOC at 850nm (invisible to optical surveillance, zero RF). Fallback: pre-computed consensus routes with probabilistic handoff zones.

### 2.3 Electronic Warfare Resilience

GPS anomaly detection within 2 seconds via VIO/terrain nav/GPS consistency check. Autonomous transition to VIO-primary on spoofing detection. S-band LPI/LPD with HAVE QUICK II frequency hopping. Directional antennas. Atomic oscillator timing backup.

---

## VOLUME III – MANAGEMENT APPROACH

### 3.1 Staffing Plan

| Role | Labor Category | Effort | Rate | CLIN |
|------|---------------|--------|------|------|
| Principal Investigator / Architect | Senior Engineer | 100% | $145/hr | 0001 |
| Swarm AI/ML Engineer | Mid-Level Engineer | 100% | $110/hr | 0001 |
| Edge Systems Engineer | Entry-Level Engineer | 100% | $82/hr | 0001 |
| Cloud Integration Engineer | Entry-Level Engineer | 100% | $82/hr | 0001 |
| T&E Engineer | Entry-Level Engineer | 100% | $82/hr | 0001 |
| Program Manager | PM | 50% | $130/hr | 0002 |
| Program Support Analyst | Admin | 50% | $78/hr | 0002 |

### 3.2 Risk Management

| Risk | Mitigation |
|------|-----------|
| EMCON coordination protocol | Pre-validated FSOC components (TRL 6+) |
| Hardware delivery timeline | Parallel procurement at award |
| AWS cost overrun | Reserved Instances + Cost Explorer alerts at 80% |
| EW test realism | Early coordination with Government test range |

---

## VOLUME IV – COST PROPOSAL & BASIS OF ESTIMATE

### 4.1 BOE Methodology

Analogous estimating anchored to completed AWS-native defense software programs. Bottom-up WBS decomposition. Rates are fully loaded composite (direct labor + fringe + OH + G&A). 1,920 annual hours/FTE (48-week productive year).

### 4.2 CLIN 0001 – Engineering Labor

| Position | Hours | Rate/Hr | Annual Cost |
|----------|-------|---------|-------------|
| Principal Investigator (Senior) | 1,920 | $145.00 | $278,400 |
| Swarm AI/ML Engineer (Mid) | 1,920 | $110.00 | $211,200 |
| Edge Systems Engineer (Entry) | 1,920 | $82.00 | $157,440 |
| Cloud Integration Engineer (Entry) | 1,920 | $82.00 | $157,440 |
| T&E Engineer (Entry) | 1,920 | $82.00 | $157,440 |
| **CLIN 0001 TOTAL** | **9,600** | | **$961,920** |

### 4.3 CLIN 0002 – Program Management

| Position | Hours | Rate/Hr | Annual Cost |
|----------|-------|---------|-------------|
| Program Manager (0.5 FTE) | 960 | $130.00 | $124,800 |
| Program Support Analyst (0.5 FTE) | 960 | $78.00 | $74,880 |
| **CLIN 0002 TOTAL** | **1,920** | | **$199,680** |

### 4.4 CLIN 0003 – Other Direct Costs

| Item | Cost |
|------|------|
| AWS Service Costs (dev/test) | $5,500 |
| Travel – Site Visits (2 trips) | $3,200 |
| Travel – Test Range (2 events) | $3,800 |
| Software Licenses & Tools | $1,500 |
| Contingency (5%) | $1,000 |
| **CLIN 0003 TOTAL** | **$15,000** |

### 4.5 Total Contract Cost Summary – Base Period

| Cost Element | CLIN 0001 | CLIN 0002 | CLIN 0003 | Total |
|-------------|-----------|-----------|-----------|-------|
| Direct Labor | $961,920 | $199,680 | — | $1,161,600 |
| ODC | — | — | $15,000 | $15,000 |
| Overhead (60% of Labor) | $577,152 | $119,808 | — | $696,960 |
| G&A (12%) | $115,430 | $38,338 | $1,800 | $155,568 |
| Fee (7% CPFF) | $115,761 | $24,981 | $1,071 | $141,813 |
| **TOTAL** | **$1,770,263** | **$382,807** | **$17,871** | **$2,847,400** |
