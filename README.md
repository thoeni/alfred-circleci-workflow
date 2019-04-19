# Alfred CircleCI Workflow

Show CircleCI Build Status, watches for builds to complete and prompts the result.

<img src="./image1.png" width="400"/>
<img src="./image2.png" width="400"/>
<img src="./image3.png" width="400"/>

## Install

Download and double-click.

- https://github.com/thoeni/alfred-circleci-workflow/releases/latest

## Usage

Register CircleCI Token.

- `apikey` saved in your `Keychain Access.app`

```
circleci token <XXXX>
```

Search and show build status.

* `circlef`: CircleCI find/filter, search by default amongst the last 30 builds, and applies the filter if provided.

	Example:
   - `circlef`
   - `circlef master`

* `circler`: CircleCI repo, search the given username and repo pair, by default amongst the last 30 builds for that repo, and applies the filter if provided.

	Example:
	- `circler thoeni alfred-circleci-workflow`
	- `circler thoeni alfred-circleci-workflow master`
	
* `circlew`: CircleCI watch, search by default amongst the last 30 builds, and applies the filter if provided, and starts polling for the result of the chosen job.

	Example:
	- `circlew alfred-circleci-workflow`
	
	Notes: This job polls every 5 seconds, with a configurable timeout of 15
	minutes. The timeout can be set in the Alfred workflow by providing a
	parameter `wt`, for example for a "1 minute" timeout, the parameter would be
	`-wt 1m`; for a 30 seconds timeout, it would be `-wt 30s`.