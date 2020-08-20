Feature: Reconcile ClusterDuckType

    Scenario Outline: Reconciling Smoke Test with generation <generation>.

        Given the following objects (from file):
            | file                      | generation   |
            | config/smoke/initial.yaml | <generation> |

        And a ClusterDuckType reconciler

        When reconciling "zhangchas.duck.knative.dev"

        Then expect status updates (from file):
            | file                      | generation   |
            | config/smoke/updated.yaml | <generation> |

        Examples:
            | generation |
            | 0          |
            | 1          |
            | 2          |
