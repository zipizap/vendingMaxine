# VendingMaxine

In the backend, Platform-engs automate platform-products provisioning

In the frontend Developers are offered those platform-products, and can self-service the ones they want. Anytime.

The platform-product will be provisioned and managed in the background, by the automation maintained by the Platform-engs.


A focused internal-developer-portal, to support teams getting the best out of the platform. Simple ;)


NOTE: not a typo, a distinctive name ;)


# User Manual


## User Workflow:

- First, the user creates a new **Collection**. 
- Then user starts **Collection Edit**, where is offered **Products** from the **Catalog**, and from which should select the products he wants.
- After user selects the products, they will be provisioned in the background - follow the progress with **Collection State**, which eventually finishes to "Completed" or "Failed".
- The user can re-start another  **Collection Edit** to change the selection of products and get them provisioned.
- If a Collection Edit "Failed", it can be re-tried by the user. If necessary a PlatformEng can download a **Collection Replayable** to inspect logs and replay the collection-edit locally for troubleshooting.

## Concepts

- **Products**: represents a *generic resource* (ex: postgresql instance)
- **Catalog**: list of *all offered products* 
- **Collection**: list of *user selected products*
-- **Collection Edit**: add/rm/upd selected products - after selection is made, the selected products will be provisioned in the background
-- **Collection State**: "Running" > "Completed" or "Failed" - *state of provisioning* of selected products 
-- **Collection Replayable**: Troubleshoot last "Failed" Collection Edit, by downloading logs and files allowing to replay the collection-edit locally 

NOTE: Processing Engines need to be carefully engineered, to be do unsupervised/atuomated provisioning changes in a resilient and idempotent way. They should be thoroughly tested during development, so that when they become generally-available to users they are sturdy and reliable (only fail on rare exceptional situations). The Processing Engines are a critical piece, their quality translates directly to user satisfaction, and sets the overall success of the vendingMaxine.

# PlatformEng Manual


## PlatformEng Workflow:

- First-time-install the VendingMaxine (see [section Installation](#Installation))
- TOREVIEW: Configure the Schema and ProcessingEngines that will present/provision the products (see [section Configuration](#Configuration)). Test provisioning works as expected. Test it thorougly, really.
- Make it available for the users
- Troubleshoot "Failed" Collection-Edit's, using Collection-Replayables to check logs and replay locally the execution
- Maintain frequent backups of database (default `./sqlite.db` file)


## Additional internal Concepts

These internal concepts extend the previous [Concepts](#Concepts) section, going into deeper details relevant for maintenance

- **Catalog**: internally is composed by **Schema** (presents user all offered products) + **ProcessingEngines** (after user selects products, it provisions them)
-- **Schema**: json-schema that creates html-form presented to user containg all offered products. From all offered products, the user can select which products wants provisioned   
-- **ProcessingEngines**: the binaries that make the provisioning of the selected products, in background. Triggered by **Collection Edit**
-- **jsonInput**, **jsonOutput**: during **Collection Edit**, these are the internal representation of previous-user-selection (jsonInput) and next-user-selection (jsonOutput)
-- **Catalog Renewal**: renew Schema + ProcessingEngines to change Products of the Catalog. Schema-renewal to change products offered, PEs-renewal so they can auto-"renew" jsonOutput on-the-fly. 




### Schema

Its the same as already documented in https://github.com/json-editor/json-editor  (dont miss also https://pmk65.github.io/jedemov2/dist/demo.html )

#### Special Additions

##### property with options.xtra_Immutable: true

Tl;dr: after property is set for the first-time, it becomes read-only and can't be changed anymore. Immutable.

In the schema, if a property contains `options.xtra_Immutable: true`, then:
- the first time this property is shown to the user, its possible to set its value
- the next times this property is shown to the user, as it already has a value, then it will be shown as "immutable" and the user cannot edit its value anymore.

This option was inspired by some terraform arguments that cannot be changed once set, as "Changing this forces a new resource to be created." ;)


### Design sketches


#### CatalogDir fs-structure

```
.../catalog-mycatalog-1-2-3/                           CatalogDir    (name: "mycatalog-1-2-3", CatalogDirBasename: "catalog-mycatalog-1-2-3")
    bin/
      internal/
        bash, busybox + busybox-symlinks, jq

      init.selfContainedEnv.source                     (for bash not sh, internally used by scripts, usefull for manual debugging)

      CollectionEdit.Launch.sh    <CollectionEditWorkdir>       (used by vendingmaxine to execute processingEngines on CollectionEdit)
      CollectionEdit.Replay.sh    <CollectionEditWorkdir>       (used by platformEng to manually troubleshoot/replay processingEngines execution)

    Schema.yaml                                        Schema.yaml will generate Schema.json. If Schema.yaml does not exist, then Schema.json must exist
    Schema.json                                        Schema.json is user-or-auto-generated, but will always exist
    00100.resource-group.Renewer_v0.0.1.sh
    00199.resource-group.Procesor.sh
    <[PE_EXECUTABLE_BINARIES](#PE_EXECUTABLE_BINARIES)>  <[PE_CONFIG_JSON_FILE](#PE_CONFIG_JSON_FILE)>

```


#### CollectionEditWorkdir fs-structure

```
.../collection-name.20230608-234202[.replay]/       CollectionEditWorkdir       (CollectionEditWorkdirBasename: "collection-name.20230608-234202")
      CollectionEditFiles/
          Schema.yaml                      Schema.yaml might not exist if user supplies Schema.json
          Schema.json
          JsonInput.json
          JsonOutput.orig.json
          JsonOutput.json                  << can be changed for renewal
          PeConfig.json                    [PE_CONFIG_JSON_FILE](#PE_CONFIG_JSON_FILE)
                  {
                    "catalog": {
                      "name": "default-0-0-0"
                    },
                    "collection": {
                      "name": "sample-col",
                      "previousState": "Completed",     << "Running" > "Completed" or "Failed"
                      "previousErrorStr": ""            << empty string, or when state=="Failed" an error-description 
                    },
                    "collection-edit": {
                      "schemaFilepath":            "./CollectionEditFiles/Schema.json",                        << read-only
                      "jsonInputFilepath":         "./CollectionEditFiles/JsonInput.json",                     << read-only
                      "jsonOutputFilepath":        "./CollectionEditFiles/JsonOutput.json"                     << read-write: on Catalog-renewal, the PEs should auto-"renew" jsonOutput on-the-fly - 
                                                                                                                in those cases change the jsonOutput-renewed in this file. Otherwise leave file unchanged
                    }
                  }
          TrackLogs/                            PeTracklogsParentdir
            00199.resource-group.Procesor.sh/    a_PeTracklogsdir
              Schema.yaml
              Schema.json
              JsonInput.json
              JsonOutput.beginning.json
              JsonOutput.json
              StdoutStderr.txt
              ExitCode.txt
    <otherFilesCreatedByPeAndAvailableToNextPe>
    <otherFilesCreatedByPeAndAvailableToNextPe>

```

#### CollectionReplayableDir fs-structure

```
.../replayable.collection-name.20230608-234202/      CollectionReplayableDir
     /catalog-mycatalog-1-2-3/                           CatalogDir    (name: "mycatalog-1-2-3", CatalogDirBasename: "catalog-mycatalog-1-2-3")
     /collection-name.20230608-234202/                   CollectionEditWorkdir  (CollectionEditWorkdirBasename: "collection-name.20230608-234202")
     README.txt
       # Replay command
       `.../xxxx/catalog-mycatalog-1-2-3/

   [ /collection-name.yyyyyyyy-yyyyyy.replay/ ]
```





#### PE_EXECUTABLE_BINARIES

##### PROCESSING ENGINE peculiarities

- executable (+x) file (or symlink to it), existing in `<CatalogDir>`
- shebang: `#!/usr/bin/env bash`
- ARG1 is <[PE_CONFIG_JSON_FILE](#PE_CONFIG_JSON_FILE)>, which will be parsed into variables (including `${JsonOutputFilepath}` which can be changed to do renewals)
- env PATH is restricted to self-contained environment of `<CatalogDir>/bin` `<CatalogDir>/bin/internal`
- env HOME is overriden to HOME="${CollectionEditWorkdir}" . This is intended to help in special cases when config files need to be placed in $HOME (like ~/.gitconfig with https authentication tokens)
- all ProcessinEngines (as this one) are executed in same PWD=CollectionEditWorkdir, so they can read/create files to other PEs run before/afterwards
- env-vars: 
  - CatalogName                     ex: 'mycatalog-1-2-3'
  - CatalogDir                      ex: '/tmp/zxkc984nas/mycatalog-1-2-3'
  - CollectionName                  ex: 'my-collection'
  - CollectionEditWorkdir           ex: '/tmp/tmp.HPjdcJ/sample-col.20230609-143317'
  - PE_CONFIG_JSON_FILE             ex: '/tmp/tmp.HPjdcJ/sample-col.20230609-143317/CollectionEditFiles/PeConfig.json'
  - SchemaFilepath                  ex: '/tmp/tmp.HPjdcJ/sample-col.20230609-143317/CollectionEditFiles/Schema.json
  - JsonInputFilepath               ex: '/tmp/tmp.HPjdcJ/sample-col.20230609-143317/CollectionEditFiles/JsonInput.json'
  - JsonOutputFilepath              ex: '/tmp/tmp.HPjdcJ/sample-col.20230609-143317/CollectionEditFiles/JsonOutput.json'
  - OLD_HOME                        ex: '/home/ubuntu'
  - HOME                            ex: '/tmp/tmp.HPjdcJ/sample-col.20230609-143317', same as CollectionEditWorkdir
  - inherits env-vars of vendingMaxine process


##### PE_CONFIG_JSON_FILE 

TODO


### ProcessingEngine

There are 2 types of ProcessingEngines (PEs):
- `Renewer_v0.1.2` which renews the jsonOutput from a previous-older-version to the new version v0.1.2. Aka Renewer-PE 
- `Processor` which should read jsonOutput and provision resource changes (create, update, delete). Aka Processor-PE


TODO: review this
      Executable files found in subdir `?????` are called ProcessingEngines (PEs). 

      When a **Collection Edit** is initiated, all the PEs are executed in sequence, in alphabetical order.  If any PE returns exit-code !=0 the sequence is interrupted and remaining PEs are not executed.

      When a PE is executed:
      - it's given 1 argument <[PE_CONFIG_JSON_FILE](#PE_CONFIG_JSON_FILE)> (ex: "/tmp/123456789.PE_CONFIG.json"). Details in section [PE_CONFIG_JSON_FILE](#PE_CONFIG_JSON_FILE)

      - it should read the PE_CONFIG_JSON_FILE::

## Instalation
TODO: with "./sqlite.db" in permanent storage


## Configuration
TODO: Schema and ProcessingEngines config, first-time and future Catalog-Renewal




# Development
## DONE && TODO

- [x] collection package and objects: db-persistence: gorm + sqlite
- [x] collection package and objects: Collection, ColSelection, ProcessingEngineRunner, ProcessingEngine, Schema 
- [x] collection package and objects: tests
- [ ] collection package and objects: logging via slog (zap)

- [x] Web API: echo framework, routes, handlers, structure
- [x] Web API: tests
- [x] Web API: /swagger/
- [x] Web API: structured logging

- [x] Web Frontend
  - [x] serve static content from web/static/*
    - [x] move js and css from CDNs to /static (noticeably speeds-up page loads)
  - [x] web /
  - [x] web /collections
    - [x] auto-refresh every 5secs
    - [x] button "new collection", show result error if exists
    - [x] buttons "edit collection" to /collections/{collectionname}
  - [x] web /collections/{collectionname}
    - [x] js+css for json-editor
    - [x] make it work
    - [ ] ? "Cancel" button to close tab
  - [x] web /schema
    - [x] Ace editor + save button

- [ ] Devise and implement strategy for ProcessingEngine execution env/args and jsonOutput-renewal
  - [x] Analysed and resumed catalog-renewal: schema-renewal requires mandatory changes on jsonInput (diagram)
  - [x] Ideate self-contained reproducible environment for processing-engines execution: locally reproduce execution of processing-engines for a collection
    - [x] Implement shell part
    - [x] Adapt golang code and remove old-tests
      - [x] remove old-tests
      - [x] Adapt ProcessingEngineRunner.run: only run a Launcher.sh

  - [x] On Collection-Edit:
    - [x] call CatalogDir/bin/internal/bash -c "CatalogDir/bin/internal/CollectionEdit.Launch.sh <JsonInputFilepath> <JsonOutputFilepath> <JsonSchemaFilepath> <CollectionInfoJsonFilepath>", which should:
      - [x] set -efu
      - [x] source CatalogDir/bin/init.selfContainedEnv.source
      - [x] export CollectionEditWorkdir=....
      - [x] cr CollectionEditWorkdir + CollectionEditFiles/*.* files
      - [x] cd CollectionEditWorkdir
      - [x] for each a_PE in order: execute a_PE <PE_CONFIG_JSON_FILE> and save TrackLogs
    - [x] save Per.CollectionEditWorkdirTgz []byte (blob)
    - [x] set ProcessingEngineRunner state with result


  - [ ] CollectionEditWorkdir as launch argument and final logs capture
      [x] golang
        [x] create and prepare CollectionEditWorkdir
        [x] call CollectionEdit.Launch.sh   <CollectionEditWorkdir>
        [x] save CollectionEditWorkdirTgzBlobId
      [x] shell
        [x] upd CollectionEdit.Launch.sh    <CollectionEditWorkdir>
        [x] upd CollectionEdit.Replay.sh    <CollectionEditWorkdir>
      [x] golang
        [x] web: Failed-message-with-link to download replayableTgz (dir containing: the-catalog/ + last-CollectionEditWorkdir/ + README.txt )
        [x] api endpoint to download collectionReplayableDirTgz    GET /api/v1/collections/:collection-name/replayable
        [x] Facilitator to download collectionReplayableDirTgz
        [x] funcs to download collectionReplayableDirTgz
        [x] replay tgz/untgz support symlinks 
      [ ] "feel the speed" - verify its usefull
        [x] remake catalog-default, a bit more realistic
          [x] fix bug with set -x in .sh scripts
          [x] pre-requisites
            [x] terraform first-time setup
              [x] azurecloud
                [x] create SPN vendingMaxine with:
                  [x] RBAC permissions:
                    [x] "Owner" on Subcription   (cr resources and roleassignments)
                    [x] "Key Vault Administrator" on Subcription (manage keyvault keys/secrets/certs)
                  [x] MicrosoftGraph application-roles: 
                    [x] "Application.ReadWrite.All" (create apps/spns)
                    [x] "Group.ReadWrite.All" (create groups)
                    [x] "User.Read.All" (read all users, for roleassignments)
                [x] create vendingmaxine-rg resource-group
                [x] create tfstate3123 storage-account, with container tfstates
              [x] local-dev: setup env-vars for azurerm
          [x] catalog PEs:
            [x] + bin/terraform
            [x] + 00100.Terraform-ProviderBackend.sh: check required env-vars, create provider.azurerm.tf + backend.azurerm.tf
            [x] + 09900.Terraform-InitPlanApply.sh: terraform init+plan+apply
            [x] + <CatalogDir>/mytfmodule

        [ ] schema.yaml as alternative to schema.json
          [x] go: If schema.yaml exists, convert it to schema.json (overwrite). If schema.yaml does not exist then schema.json must exist. Finally use schema.json. Schema.json will always exist in the end.
          [ ] go: yaml2json enhancements  (~/tmp/tmp.20230826172226.IHKd)
            [x] Schema.json: resume different scenarios that can happen, with relevant presentation-fields and data-fields
              [x] composite types
                [x] object
                [x] array
              [x] basic types
                [x] string
                [x] integer
                [x] number
                [x] boolean
              [ ] review and add json-editor features
            [ ] Schema.yaml: enhancements to simpler and powerfull yaml structure to express presentation-fiels and data-fields, and how they convert to Schema.json
            [ ] go: implement Schema.yaml enhancements and convertion to Schema.json
            [ ] "feel the speed" - verify its usefull ;)

        [ ] markdown descriptions: webpage js should render markdown->html in all descriptions of the schema, before using the schema in the jseditor

        [ ] je: revise options: theme, bootstrap5?  any theme but should look organized and show descriptions with htlm (bootstrap3 did it in examples)

        [ ] web collection: 
          [ ] disable bottom-button while json-editor not valid and show validation-errors bellow the button
          [ ] cancel button to go back to collections webpage

        [ ] improve json-editor config, to:
          [ ] render htlm in title/descriptions
          [ ] show lines, as in the demo (easier to follow)



        [ ] impressions
          [ ] editing catalog still feels clunky
          [ ] logs: access them also when collection "Completed"
          [ ] logs: quick-view ?in tab?


        [-] on collectionEdit-ok, allows to check logs
        [x] on collectionEdit-nok, allows to check logs with error
        [x] on collectionEdit-nok, allows to download collectionReplayableDirTgz and replay it locally, multiple times
       
      [x] property with options.xtra_Immutable: true (value can only be set once and then becomes readonly)
        [x] in web js, disable editors with that option set
          [ ] see it in action :)
        [ ] api: validate it in api

  - [ ] execute PEs on collection creation, before allowing user to edit the collection, to allow Renewer-PEs to set in jsonoutput any mandatory elements (ex: mandatory 1st storage account with xtra_Immutable params, etc)


  - [x] Secrets handling
      [x] All secrets are sourced from env-vars for now, inherited from the env-vars set to the parent-process (vendingMaxine process)
      [x] $HOME overriden to HOME="${CollectionEditWorkdir}" to help in special cases when PEs need to create $HOME config-files, like ~/.gitconfig  

  - [ ] button "Save and update collection" should only be enabled when editor.valid (all editor validations are ok)

  - [ ] Additional env-vars to be set for PEs:
      [ ] 

  - [ ] On Catalog-renewal
    - [ ] TODO

  - [ ] ? Formalize Catalog concept: Catalog = Schema+PEs from git-ref, with versions by tag
    - [ ] Catalog from git-repo
  
  - [ ] ? Multiple Catalogs?
    - [ ] Each Catalog with a name, a git-repo, and multiple versions (corresponding to multiple git-refs tags or branch-names)
    - [ ] Each Collection selects its Catalog on creation (and not changeable ever after). Each Collection keeps its Catalog-name and Catalog-version
    - [ ] Catalog-renewal:
      - [ ] Flow:
            . git: new git-ref "v0.0.2"
            . VdMx: API to "Renew collection to new version v0.0.2", per Collection. 
              This per collection migration allows:
                . beta-testing new-version on tmp-collections, during development
                . validation-testing new-versions on validating-collections, before renewing all user-collectinos
                . renew user collections in a granular controlled way: just one, then a few, then all of them
              Git refs similar to this:
                https://developer.hashicorp.com/terraform/language/modules/sources#selecting-a-revision
                https://git-scm.com/book/en/v2/Git-Tools-Revision-Selection#_single_revisions
              Git shallow clone (--depth=1):
                https://developer.hashicorp.com/terraform/language/modules/sources#shallow-clone


  - [x] Readme documented the high-level design of concepts and operations, including definition of PE_CONFIG_JSON_FILE 

- [ ] collections.html: create collections using available catalogs other than "default-0-0-0"

- [ ] States on objects: review, fix and add tests

- [ ] json-editor: mark json-input fields to become "set-one-time-and-then-readonly-forever"

- [ ] web collections, show last-update time in table for each collection 

- [x] (improv) Execute colection-edit-save in background instead of waiting synchronously (web::putCollectionEditSave http-return "200ok" and let inner-stuff run in background)

- [ ] ?? Catalog Renewal: procEng+schema (api copy-from-git-ref?) 

- [ ] ? webpage Collection PEs: log, status, files 

- [ ] verify schema update triggers update of all collections

- [ ] ? cancel buttons?

- [ ] delete collection via api endpoing (not via webpage for now)

- [ ] Export-all, restore-all: its all in db, so copy/restore db and restart (there is config-option to set db-file)

- [ ] ?? review states "Running"... ??

- [ ] ? rename VendingMaxine to more fitting name
      All these are taken:  WonderHub, WonderWorks, Wonderland, WonderStation, WonderPortal, WonderPort, 
      These are not taken: WonderPortal



## Local setup

### Requirements

- echo-swagger : `go install github.com/swaggo/swag/cmd/swag@latest`
