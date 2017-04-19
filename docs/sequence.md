# Sequence Diagram

@startuml
participant GitHub
participant Acid
participant AcidJS
participant Kubernetes

GitHub->Acid: push webhook
Acid<->GitHub: git clone repo
Acid->Kubernetes: Get configuration for this project
Acid->Acid: Load acid.js
Acid->AcidJS: Run
AcidJS->Kubernetes: Create settings configmap
AcidJS->Kubernetes: Run job in pod
AcidJS->Kubernetes: Run other job...
Kubernetes->AcidJS: Success
AcidJS->Acid: All done
Acid->GitHub: Success
@enduml
