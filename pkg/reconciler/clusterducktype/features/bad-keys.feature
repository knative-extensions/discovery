Feature: Reconcile ClusterDuckType

    Scenario Outline: Reconciling <key> causes <result>.

        Given the following objects:
            """
            """
        And a ClusterDuckType reconciler
        When reconciling "<key>"
        Then expect <result>

        Examples:
            | key            | result  |
            | too/many/parts | nothing |
            | foo/not-found  | nothing |
