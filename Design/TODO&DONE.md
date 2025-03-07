## Sparse notes

- proj survival depends on splitting small&easy tasks

- better simple and clear (even if longer) than complex and terse

- ObjA, ObjAA, ObjAAA:   
  ObjA can call ObjAA which can call ObjAAA.   
  ObjA cannot call ObjAAA directly. ObjAAA cannot call ObjAA or ObjA.   
  This way, the dependencies are clear and the code is easier to maintain and understand.

- userlogin: for now just have a stub with username "admin", email "admin@local.local", uniqueID "00000000-0000-0000-0000-000000000000"

- Drawio diagram: prefer web version (desktop-app requires ?root-suid? wtf?)  
  https://app.diagrams.net/




## TODO


- Tzzz) gorm does not support fields of type []string (error "unsupported data type: &[]")
  - Create new 

- Tzzz) gorm::main.go: ReLoading from db the AutoBus records (Alice should only be in Bus2)
  Not working, see how to properly delete an element from a slice in gorm 

- Tzzz) start adding webpages

- Tzzz) add description field to the following types.
  It should initially be set by constructor, with getter/setter methods GetDescription() SetDescription()
  - Collection (and underlying DbCollection)
  - ColRevision (and DbColRevision)

- Tzzz) When RevState is "ProvisiningFailed". there is no next-state possible. How to solve this?


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


