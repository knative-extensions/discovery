Feature: Reconcile ClusterDuckType

    Scenario Outline: Reconciling Smoke Test for Duck <dkv>.

        Given the following objects (from file):
            | file                      | generation   |
            | config/smoke/initial.yaml | <generation> |

        And a <dkv> metareconciler

        When reconciling "zhangchas.duck.knative.dev"

        Then expect <result>

        Examples:
            | dkv                                | result    |
            | "addressables.duck.knative.dev/v1" | nothing   |
            | "zhangchas.duck.knative.dev/v1"    | 1 controllers |
