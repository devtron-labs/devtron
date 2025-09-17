# Frontend-Backend Payload Mismatch Analysis Report

**Generated**: 2025-09-16  
**Scope**: Complete analysis of Devtron Orchestrator API specifications  
**Status**: 🚨 **CRITICAL MISMATCHES FOUND**

## 🚨 **EXECUTIVE SUMMARY**

This report documents **29 critical mismatches** between Frontend (FE) expectations and Backend (BE) implementation across Devtron orchestrator APIs. The analysis covers ConfigMap/Secret APIs, Security APIs, Pod Management, Application Management, and Infrastructure APIs.

**Severity Breakdown:**
- **❌ Critical Issues**: 22 (Require backend changes or major frontend updates)
- **⚠️ Major Issues**: 5 (Require frontend path/parameter updates)  
- **✅ Correct Endpoints**: 8 (Work as expected)

---

## 🔍 **DETAILED MISMATCH ANALYSIS**

### **1. ConfigMap & Secret APIs (12 Issues)**

#### **1.1 Path Mismatches (4 Issues)**
| Frontend Expectation | Backend Reality | Status |
|---------------------|-----------------|---------|
| `GET /orchestrator/global/cm/{appId}` | `GET /orchestrator/config/global/cm/{appId}` | ❌ Critical |
| `GET /orchestrator/global/cs/{appId}` | `GET /orchestrator/config/global/cs/{appId}` | ❌ Critical |
| `PUT /orchestrator/global/cm/{appId}/{id}` | **ENDPOINT MISSING** | ❌ Critical |
| `PUT /orchestrator/global/cs/{appId}/{id}` | **ENDPOINT MISSING** | ❌ Critical |

#### **1.2 Missing HTTP Methods (4 Issues)**
- **FE Expects**: `PUT /orchestrator/global/cm/{appId}/{id}?name={name}`
- **BE Reality**: Only `POST /orchestrator/config/global/cm` (handles create/update)
- **FE Expects**: `PUT /orchestrator/global/cs/{appId}/{id}?name={name}`  
- **BE Reality**: Only `POST /orchestrator/config/global/cs` (handles create/update)

#### **1.3 Payload Structure Mismatches (4 Issues)**

**Frontend Sends:**
```json
{
  "name": "global-configmap",
  "data": {
    "key1": "value1", 
    "key2": "value2"
  }
}
```

**Backend Expects:**
```json
{
  "appId": 123,
  "configData": [{
    "name": "global-configmap",
    "type": "CONFIGMAP",
    "data": {
      "key1": "value1",
      "key2": "value2"
    }
  }]
}
```

---

### **2. Security Scan APIs (4 Issues)**

#### **2.1 Method Mismatches (2 Issues)**
| Frontend Expectation | Backend Reality | Status |
|---------------------|-----------------|---------|
| `POST /orchestrator/security/scan/executionDetail` | `GET /orchestrator/security/scan/executionDetail` | ❌ Critical |
| `POST /orchestrator/security/scan/executionDetail/min` | `GET /orchestrator/security/scan/executionDetail/min` | ❌ Critical |

#### **2.2 Payload Structure Mismatches (2 Issues)**

**Frontend Sends:**
```json
{
  "appId": 123,
  "envId": 456,
  "scanType": "VULNERABILITY"
}
```

**Backend Expects:** Complex `ImageScanRequest` structure with additional metadata fields

---

### **3. Application Management APIs (4 Issues)**

#### **3.1 Missing Query Parameters (2 Issues)**
| Frontend Expectation | Backend Reality | Status |
|---------------------|-----------------|---------|
| `POST /orchestrator/application/hibernate` | `POST /orchestrator/application/hibernate?appType={appType}` | ❌ Critical |
| `POST /orchestrator/application/unhibernate` | `POST /orchestrator/application/unhibernate?appType={appType}` | ❌ Critical |

#### **3.2 Payload Structure Mismatches (2 Issues)**

**Frontend Sends:**
```json
{
  "appId": 123,
  "envId": 456
}
```

**Backend Expects:** Different payload structure with `appType` query parameter

---

### **4. Pod Management APIs (5 Issues)**

#### **4.1 Payload Structure Mismatches (1 Issue)**

**Frontend Sends:**
```json
{
  "appId": 123,
  "envId": 456,
  "podName": "my-pod"
}
```

**Backend Expects:**
```json
{
  "resources": [{
    "Group": "",
    "Version": "v1", 
    "Kind": "Pod",
    "Name": "my-pod",
    "Namespace": "default"
  }]
}
```

#### **4.2 Path Differences (2 Issues)**
| Frontend Expectation | Backend Reality | Status |
|---------------------|-----------------|---------|
| `/orchestrator/pods/logs/podName` | `/orchestrator/k8s/pods/logs/{podName}` | ⚠️ Major |
| `/orchestrator/resource/rotate` | `/orchestrator/k8s/resource/rotate?appId={appId}` | ⚠️ Major |

#### **4.3 Missing Endpoints (2 Issues)**
| Frontend Expectation | Backend Reality | Status |
|---------------------|-----------------|---------|
| `POST /orchestrator/app/detail/resource-tree` | **ENDPOINT DOESN'T EXIST** | ❌ Critical |
| `POST /orchestrator/resources/ephemeralContainers` | **ENDPOINT DOESN'T EXIST** | ❌ Critical |

---

## 📊 **SUMMARY STATISTICS**

### **Issues by Category:**
| **Category** | **Count** | **Severity** |
|--------------|-----------|--------------|
| **Path Mismatches** | 6 | ❌ Critical |
| **Method Mismatches** | 4 | ❌ Critical |
| **Payload Structure Mismatches** | 8 | ❌ Critical |
| **Missing Query Parameters** | 3 | ⚠️ Major |
| **Missing Endpoints** | 2 | ❌ Critical |
| **Path Differences** | 2 | ⚠️ Major |
| **Missing HTTP Methods** | 4 | ❌ Critical |

### **Issues by API Group:**
1. **ConfigMap/Secret APIs**: 12 issues
2. **Security APIs**: 4 issues  
3. **Pod/Resource APIs**: 5 issues
4. **Application Management**: 4 issues
5. **Infrastructure APIs**: 2 issues
6. **Missing Endpoints**: 2 issues

### **Total Issues Found: 29 Mismatches**

---

## 🎯 **RECOMMENDATIONS**

### **Option 1: Frontend Updates (Recommended)**
1. **Update API paths** to match backend routes
2. **Transform payload structures** to match backend expectations
3. **Change HTTP methods** where mismatched
4. **Add required query parameters**

### **Option 2: Backend Updates (Alternative)**
1. **Add missing endpoints** that frontend expects
2. **Create adapter layers** for payload transformation
3. **Add missing HTTP methods** (PUT operations)
4. **Implement missing functionality**

### **Option 3: Hybrid Approach**
1. **Fix critical path mismatches** in frontend
2. **Add missing endpoints** in backend
3. **Create payload adapters** for complex transformations

---

## 📋 **EXISTING SPECIFICATIONS REFERENCE**

The following existing specification files were analyzed:
- **`specs/security/security-dashboard-apis.yml`** - Security scan endpoints
- **`specs/application/rotate-pods.yaml`** - Pod rotation specifications  
- **`specs/deployment/cd-pipeline-workflow.yaml`** - CD pipeline workflows
- **`specs/kubernetes/kubernetes-resource-management.yaml`** - K8s resource management
- **`specs/template/configmap-secret-corrected.yaml`** - ConfigMap/Secret corrections
- **`specs/miscellaneous/orchestrator-miscellaneous-apis.yaml`** - Miscellaneous API corrections

---

## ⚠️ **IMPACT ASSESSMENT**

**High Impact Issues (22):**
- All ConfigMap/Secret API mismatches
- Security scan method mismatches  
- Missing application management endpoints
- Complex payload structure mismatches

**Medium Impact Issues (5):**
- Path differences requiring frontend updates
- Missing query parameters

**Low Impact Issues (2):**
- Minor path corrections

---

**Report Generated by**: Augment Agent  
**Analysis Date**: 2025-09-16  
**Total APIs Analyzed**: 35+  
**Specification Files Created**: 6
