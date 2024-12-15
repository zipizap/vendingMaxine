# vendingMaxine
Developers can self-service products, and platform/sres automate products lifecycle

NOTE: vendingMaxing is not a typo, but a way to add "uniqueness" to the name


## Concepts for the user

- **Products**: represents a *generic resource* (ex: postgresql instance)
- **Catalog**: list of *all offered products* 
- **Collection**: list of *user selected products* (ex: mypostgresql, mykeyvault)
- **Collection Edit**: add/rm/upd selected products 
  - On each **Collection Edit**, the selection is  changed by the user, and at the selected products will be provisioned in the background
  - Each **Collection Edit** is represented by an increasing  **CollectionRevision**, and so the **CollectionRevision(s)** track the evolution of the **Collection** over time.
  - The **CollectionRevision** stores a **RevState** showing the state of selected products

NOTE: Processing Engines need to be carefully engineered, to be do unsupervised/atuomated provisioning changes in a resilient and idempotent way. They should be thoroughly tested during development, so that when they become generally-available to users they are sturdy and reliable (only fail on rare exceptional situations). The Processing Engines are a critical piece, their quality translates directly to user satisfaction, and sets the overall success of the vendingMaxine.


## Webpages

```
         (1)
WebHome -->-- WebCollections 
                  |     
                  +-->-- WebCollectionNew  
                  |
                  +-->-- WebCollectionEdit
                  |
                  +-->-- WebCollectionDashboard


(1) user logs.in

```


Public mainpage:  **WebHome**
- user login
- quick-demo with video or images or explanation
- link to public docs

Once user logs in: **WebCollections**
- Table showing all collections belonging to the user
  - One collection per line, with collection name (linked to **WebCollectionDashboard**) and resume of last CollectionRevision (modification-time, RevState, date, modified-by user)
  - If RevState is "Completed" or "Failed" (but not "Running"), then show a button "Edit Collection" (leading to **WebCollectionEdit**)
- Button "Create collection" (leading to **WebCollectionNew**)

When user wants to create a new collection: **WebCollectionNew**
- Ask user:
  - CollectionName
- Button "Create", which should send a request to the server to create the new collection and get back a response with Completed/Failed
  - Response with Completed: should send back and show user collectionID and CollectionName
  - Response with Failed: should send back error and show user error
- Button "Cancel" (leading to **WebCollections**)

When user wants to edit a collection: **WebCollectionEdit**
- TODO: Provide detailed instructions for editing a collection
- Button "Save"
  - ...
- Button "Cancel"
  - ...

The **WebCollectionDashboard**
- Info from last **CollectionRevision**: last-modification date, user, RevState
- History of previous **CollectionRevision(s)** with dates, user, RevState - in descending order (latest on top)


## Concepts

### CollectionRevision
- **RevID**
- **RevStates[RevState]**: [0] is oldest, [len-1] is latest
  - **State**: one of
    - `UserEditing`: The collection edit has started, but not yet completed 
    - `UserCancel`: The collection edit has started, and then cancelled
    - `ProvisionRunning`: The provisioning of the selected products is currently in progress.
    - `ProvisionCompleted`: The provisioning of the selected products has finished successfully.
    - `ProvisionFailed`: The provisioning of the selected products has encountered an error and did not complete successfully.
    ```
        UserEditing   ------->  UserCancel     
            |
            v
        ProvisionRunning  ---> ProvisionFailed
            |
            v
        ProvisionCompleted
    ```
  - **DateStart**
- **RevState**: alias to the latest **RevStates[len-1]** 
- **RevLogs**: binary data (zip-file)
- **User**

## DockerImage
- All logs inside `/logs`, easy to zip-and-download at the end of execution

### Start of CollectionEdit
- **Server** creates new **CollectionRevision** filling in:
  - **RevID**
  - **User**
  - cr **RevState** and append to **RevStates[]**


### Finale of CollectionEdit
- **Server** fills in **CollectionRevision**
  - **RevLogs**: **Server** requests to **dockerimage** which replies with zip of `/logs`
  - cr **RevState** and append to **RevStates[]**



### User
  - **UserName**
  - **Email**





# Development
## Sparse notes

- Drawio diagram: prefer desktop version (https://app.diagrams.net/)
- React POC tests: https://stackblitz.com/edit/react-dwbpzf?file=src%2FGenericForm.tsx


## TODO
- rename "CollectionTemplate" with "CatalogBlueprint"
  - diagram
  - readme

## DONE