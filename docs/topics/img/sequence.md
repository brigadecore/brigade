# Sequence Diagram

@startuml
participant GitHub
participant Brigade
participant BrigadeJS
participant Kubernetes

GitHub->Brigade: push webhook
Brigade<->GitHub: git clone repo
Brigade->Kubernetes: Get configuration for this project
Brigade->Brigade: Load brigade.js
Brigade->BrigadeJS: Run
BrigadeJS->Kubernetes: Create settings configmap
BrigadeJS->Kubernetes: Run job in pod
BrigadeJS->Kubernetes: Run other job...
Kubernetes->BrigadeJS: Success
BrigadeJS->Brigade: All done
Brigade->GitHub: Success
@enduml
