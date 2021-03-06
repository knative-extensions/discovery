Feature: Reconcile ClusterDuckType

    Scenario Outline: Reconciling <key> causes <result>.

        Given the following objects:
            """
            """
        And a "addressables.duck.knative.dev/v1" metareconciler
        When reconciling "<key>"
        Then expect <result>

        Examples:
            | key            | result  |
            | too/many/parts | nothing |
            | foo/not-found  | nothing |
