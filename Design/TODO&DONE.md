## Sparse notes

- proj survival depends on splitting small&easy tasks

- better simple and clear (even if longer) than complex and terse

- userlogin: for now just have a stub with username "admin", email "admin@local.local", uniqueID "00000000-0000-0000-0000-000000000000"

- Drawio diagram: prefer web version (desktop-app requires ?root-suid? wtf?)  
  https://app.diagrams.net/




## TODO

WIP - models and dbModels foundation
  + ColRev and up:
    + check ColRev
    + integrate ColRev into Collection
  - ColRev and down:
    - check state transitions: validated and part of RevState

  - IsEditable: RevState, ColRev, Collection

  - logic: Col creates ColRev which creates RevState. What are RevState args, so that ColRev and Col set them




- T006) add description field to the following types.
  It should initially be set by constructor, with getter/setter methods GetDescription() SetDescription()
  - Collection (and underlying DbCollection)
  - ColRevision (and DbColRevision)

- Add new RevState: "Ready"
  When everything is done and complete, it should become "Ready"
    - ProvisioningCompleted should transition to "Ready"

- ?? As of now, a new collection is created with empty Col.ColRevisions[], and some methods have to check if ColRevisions[] is empty or not (some related to ColRevisionLatest).
  If we added the RevState "Ready" as the default RevState of new ColRevisions, then ColColRevisions[] would never be empty and that would simplify some code in the other methods.
  However, that would mean we could have 2 different RevState flows:
    - the CollectionEditFlow: from CollectionEditOngoing ... ProvisioningComplete -> Ready
    - a new CollectionCreatedFlow: just one state "Ready"
      And when new collection is created then Col.ColRevisions[] would be non-empty and contain this Ready state.


- When RevState is "ProvisiningFailed". there is no next-state possible. How to solve this?


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
- T002) Add a CollectionID field, and use it internally instead of CollectionName. So different users can have collections of same name but different id

- T005) helper-prefixes on ObjectsIDs, like:
          DbCollection.ID    as  <UUID>
          CollectionID()     as  "ColID-<DbCollection.ID>"
        Applied to all suitable objects
        This shuold provide a clear and consistent way to identify objects in the code/debug/logs, and avoid confusion between objects of different types.

- T004) RevStates: from map to []RevState, to allow for easy ordering of states and containement of other props specific to each state (as inputs, outputs, etc)
- T003) RevStates: no longer array, use map insted to allow free containment other props specific to each state (as inputs, outputs, etc)
- T001) rename "CollectionTemplate" with "CatalogBlueprint"
  - diagram
  - readme


