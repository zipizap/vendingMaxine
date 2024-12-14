# vendingMaxine
Developers can self-service products, and platform/sres automate products lifecycle

NOTE: vendingMaxing is not a typo, but a way to add "uniqueness" to the name


## Concepts

- **Products**: represents a *generic resource* (ex: postgresql instance)
- **Catalog**: list of *all offered products* 
- **Collection**: list of *user selected products* (ex: mypostgresql, mykeyvault)
-- **Collection Edit**: add/rm/upd selected products - after selection is made, the selected products will be provisioned in the background
-- **Collection State**: "Running" > "Completed" or "Failed" - *state of provisioning* of selected products
-- TBD **Collection Replayable**: Troubleshoot last "Failed" Collection Edit, by downloading logs and files allowing to replay the collection-edit locally 

NOTE: Processing Engines need to be carefully engineered, to be do unsupervised/atuomated provisioning changes in a resilient and idempotent way. They should be thoroughly tested during development, so that when they become generally-available to users they are sturdy and reliable (only fail on rare exceptional situations). The Processing Engines are a critical piece, their quality translates directly to user satisfaction, and sets the overall success of the vendingMaxine.


## Webpages

```
         (1)
WebHome -->-- WebCollections 
                  |     
                  +-->-- WebCollectionNew  
                  |             
                  |
                  |-->-- WebCollectionEdit


(1) user logged in

```


public mainpage:  **WebHome**
- user login
- quick-demo with video or images or explanation
- link to public docs

Once user logs in: **WebCollections**
- table showing all collections bellonging to the user
- one collection per line, with collection name (linked to **WebCollectionHistory) and last-modification CollectionState, date, modified-by user 
-- If CollectionState is "Completed" or "Failed" (but not "Running") then show a button "Edit Collection" (leading to **WebCollectionEdit**) 
- Somewhere have a button "Create collection" (leading to **WebCollectionNew**)

When user wants to create new collection: **WebCollectionNew**
- ask user:
    - collectionName
- one button "Create", which shuold send request to backend, to create collection, and get back a response with Completed/Failed
  -- response with Completed: should send back and show user collectionID and collection name
  -- response with Failed: should send back error and show user error
- one button "Cancel" (leading to **WebCollections**)

When user wants to edit collection: **WebCollectionEdit**
- TODO


**WebCollectionHistory** ?? not best name... 
- show:
-- last-modification date, user, CollectionState
-- history of previous CollectionState(s) with dates
-- Creation date and user


# Development
## Sparce notes

- Drawio diagram: prefer desktop version ( https://app.diagrams.net/ )
- react poc tests: https://stackblitz.com/edit/react-dwbpzf?file=src%2FGenericForm.tsx 