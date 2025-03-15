## Sparse notes

- proj survival depends on splitting small&easy tasks

- better simple and clear (even if longer) than complex and terse

- ObjA, ObjAA, ObjAAA:   
  ObjA can call ObjAA which can call ObjAAA.   
  ObjA cannot call ObjAAA directly. ObjAAA cannot call ObjAA or ObjA.   
  This way, the dependencies are clear and the code is easier to maintain and understand.


- Drawio diagram: prefer web version (desktop-app requires ?root-suid? wtf?)  
  https://app.diagrams.net/




## TODO

- T013) Introduce GlobalAdminGroup, GlobalReaderGroup:
  - Logic:
    . It's value should be read from the config.yaml and not saved into db.  
    . At runtime, the related functions should use it to validate access.  
    . This allows the GlobalAdminGroup/GlobalReaderGroup to be changed in a simple way, by changing config.yaml and restarting the app.  
  - improve config.yaml to read GlobalAdminGroup,GlobalReaderGroup into a global variable
  - improve the opHub methods and types, to receive in their func params the GlobalAdminGroup,GlobalReaderGroup values, to validate access 
 
- Tzzz) start adding webpages

- Tzzz) When RevState is "ErrorProvisioningFailed". there is no next-state possible. How to solve this? 
  . ? Let GlobalAdmins be able to perform some global-admin-special-operations, like forcing a transition to "ReadyToRetry" or something? 

- Tzzz) add some tests to the model and dbmodel - at least for the most-significant operations  

- T999) clean existing TOREVIEW, TBD, TODO, WIP



### TODO-future-versions?

- fix deadlock-by-silent-runner-death: if a runner gets unexpectedly killed (or server gets killed), the collection will keep the Ongoing state forever. There should be a mechanism (api call or whatever) to change such states in the db to Failed-or-appropriate, so that it can be restarted by the user

- add governance:
  - implement userlogin via DEX
  - WebCollectionNew should ask user which user/group-objid should be owners of the collection
  - Collection should store field with owners user/group-objids
  - each http request shuold validate and let current user access only the collections he is owner of (deny all other requests for collections not-owner)
  - in-some-web there should be possible to update the user/group-objids

- WebCollectionDelete + etc
- CollectionNew and CollectionDelete scripts: \CatalogBlueprint\{CollectionNew,CollectionDelete} 


## DONE

+ T012) Think how to do user-authz with Collection (Models) operations
  + Logic:
    . Assume we will have a currentUserUser and currentUserGroups that can be passed in the parameter struct, to validate access
    . We could implement a package `opHub` that separates and mediates all operations between webserver side and Models  
      This way we could centralize and control all operations required by webserver, in a separate way from the Models.
      And in this opHub it would be possible to enforce user-authz 
    . All types zzzReq should contain a field `Client CurrentClient` that will be used to validate access in the methods
    . All methods should validate access of `Client CurrentClient`, and return an error if access is denied
  + Create opHub package with the following methods:
    + CollectionsList(req *CollectionsListReq) (resp *CollectionsListResp, err error) 
    + CollectionNew(req *CollectionNewReq) (resp *CollectionNewResp, err error) 
    + DoCollectionEditOngoing(req *CollectionEditOngoingReq) (resp *CollectionEditOngoingResp, err error)
    + DoCollectionEditCancelled(req *CollectionEditCancelledReq) (resp *CollectionEditCancelledResp, err error)
    + DoCollectionEditCompleted(req *CollectionEditCompletedReq) (resp *CollectionEditCompletedResp, err error)
    + add validation of CurrentClient


+ T011) add description field to the following types.
  It should initially be set by constructor, with getter/setter methods GetDescription() SetDescription()
  - Collection (and underlying DbCollection)
  - ColRevision (and DbColRevision)


+ T010) gorm::main.go: ReLoading from db the AutoBus records (Alice should only be in Bus2)
  Not working, see how to properly delete an element from a slice in gorm 

+ T009) add DbAccessPolicyParams so it is used in all method-params that capture it (collection methods or whatever). It will be better to have a type there instead of loosen []strings
  ```
    type DbAccessPolicyParams struct {
      AdminUsers   []string
      AdminGroups  []string
      ReaderUsers  []string
      ReaderGroups []string
    }
  ``` 

+ T008) gorm does not support fields of type []string (error "unsupported data type: &[]")
  + undo DbAdminUser 
    + del DbAdminUser

  + re-implement AccessPolicy related things:
    ```
    DbCollection.DbAccessPolicy  1DbCollection-to-1DbAccessPolicy

    DbAccessPolicy.DbAccessPolicyMappings []*DbAccessPolicyMapping    1DbAccessPolicy-to-manyDbAccessPolicyMapping

    DbAccessPolicyMapping
      DbCollectionID
      Role
      User
      Group

    ```
    + Logic:   
      . AccessPolicy - just a facade with methods that return AdminUsers, AdminGroups, ReaderUsers, ReaderGroups []string, and other usefull methods
      . DbAccessPolicyMapping - a new underlying model, which will be a single table containing rules for all collections, for all roles and all users/groups 
      . AccessPolicy will call DbAccessPolicyMapping methods, which in turn will query db to return the []string and other desired data.
      . DbAccessPolicyMapping will have the following fields:
        . CollectionID
        . Role
        . User
        . Group
    + review AccessPolicy methods: use DbAccessPolicyMapping to get []string and should work




+ T007) - models and dbModels foundation
  + ColRev and up:
    + check ColRev
    + integrate ColRev into Collection
  + ColRev 
    + update state transitions ColRev-N, ColRev-N+1
      + IsValidRevStateTransition() code
      + rename func IsValidRevStateTransition()
    + ColRev funcs to review:
      + constructor of RevState shuold use the IsValidRevStateTransition()
      + ColRev.IsEditable() should use ColRev.IsValidRevStateTransition("CollectionEditOngoing") 
  + IsEditable: 
    + ColRev
    + Collection

  + Improve the logic of IsValidRevStateTransition()
    + first move all logic of IsValidRevStateTransition() out of RevState, and into ColRev - this logic should be in ColRev
      + DbRevStateNew() callers should assure isValidRevStateTransition() before calling it
      + isValidRevStateTransition(currentState, newState) 
      + DbColRevisionNew(... currRevStateName new arg)
      + review and align Collection.Logic (See ## concepts internals)
    + ColRev.IsValidRevStateTransition() shuold get the currentState either from the actual ColRev or from a previous ColRev (if the current one is nil) 
        IsValidRevStateTransition(currentState, newState) 
        If no RevState exists in this ColRev
        a) try to get the stateLatest from the previous ColRev
        b) if it also does not exist, then assume this is the first RevState of first ColRev and assume the state is "Ready"
        Maybe this logic should be moved to the ColRevision level or even to the Collection level, so that there the currState can be determined
        (from this or a previous ColRevs) and then passed into here as a parameter
        (And cleanup current hacky logic that hardcodes `currStateName = "Ready"`

  + Model.zzz methods sometimes implement logic that could be moved to the direct counterpart in dbModel.zzz methods
    It would be better to move the logic to the dbModel.zzz methods, and then call them from the Model.zzz methods
    + Collection
    + ColRevision

  + RevState
    + integration with ColRev
    + RevState code review

  + Create: GetRevStateLatestUserWhoTriggered() in ColRev and Collection (and their dbModels)
    + func (d *DbColRevision) GetRevStateLatestUserWhoTriggered() (string, error)
      + should read from the latestRevState.GetUserWhoTriggered()
    + func (o *ColRevision) GetRevStateLatestUserWhoTriggered() (string, error)
      + shuold read from `(d *DbColRevision) GetRevStateLatestUserWhoTriggered()`
    + func (d *DbCollection) GetRevStateLatestUserWhoTriggered() (string, error)
      + should read from the latestRevStateFromLastColRev.GetUserWhoTriggered()
    + func (o *Collection) GetRevStateLatestUserWhoTriggered() (string, error)
      + should read from `(d *DbCollection) GetRevStateLatestUserWhoTriggered()`

  + logic: Col creates ColRev which creates RevState. What are RevState args, so that ColRev and Col set them
  + ColRev: implement user-who-triggered to pass it down to RevState (or improve this user-identification somehow and then align to it RevState and ColRev)

  + As of now, a new collection is created with empty Col.ColRevisions[], and some methods have to check if ColRevisions[] is empty or not (some related to ColRevisionLatest).
    If we define the RevState "Ready" as the final-RevState of any ColRev that has ended correctly, then when a new collection is created, we could add an initial ColRev with RevState "Ready", that signals that a future ColRev can be created from this "Ready" to proceed
    And this way, when a new collection is created, its  Col.ColRevisions[] would be filled with an initial ColRev with RevState "Ready". This way, we would avoid the need to check if ColRevisions[] is empty or not, and we could always use ColRevisionLatest() without checking if it is nil or not.
    This would also allow to have a consistent behavior in the code, and avoid some bugs that could arise from the current behavior.



+ T006) Add new RevState: "Ready"
    + When everything is done and complete, it should become "Ready"
    + ProvisioningCompleted should transition to "Ready"


+ T002) Add a CollectionID field, and use it internally instead of CollectionName. So different users can have collections of same name but different id

+ T005) helper-prefixes on ObjectsIDs, like:
          DbCollection.ID    as  <UUID>
          CollectionID()     as  "ColID-<DbCollection.ID>"
        Applied to all suitable objects
        This shuold provide a clear and consistent way to identify objects in the code/debug/logs, and avoid confusion between objects of different types.

+ T004) RevStates: from map to []RevState, to allow for easy ordering of states and containement of other props specific to each state (as inputs, outputs, etc)

+ T003) RevStates: no longer array, use map insted to allow free containment other props specific to each state (as inputs, outputs, etc)

+ T001) rename "CollectionTemplate" with "CatalogBlueprint"
  - diagram
  - readme


