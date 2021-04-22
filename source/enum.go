package source

import (
	ghv4 "github.com/shurcooL/githubv4"
)

type PullRequestState = ghv4.PullRequestState

const (
	PullRequestStateOpen   PullRequestState = ghv4.PullRequestStateOpen
	PullRequestStateClosed PullRequestState = ghv4.PullRequestStateClosed
	PullRequestStateMerged PullRequestState = ghv4.PullRequestStateMerged
)

type IssueOrderField = ghv4.IssueOrderField

const (
	IssueOrderFieldCreatedAt IssueOrderField = ghv4.IssueOrderFieldCreatedAt
	IssueOrderFieldUpdatedAt IssueOrderField = ghv4.IssueOrderFieldUpdatedAt
	IssueOrderFieldComments  IssueOrderField = ghv4.IssueOrderFieldComments
)

type OrderDirection = ghv4.OrderDirection

const (
	OrderDirectionAsc  OrderDirection = ghv4.OrderDirectionAsc
	OrderDirectionDesc OrderDirection = ghv4.OrderDirectionDesc
)
