/*
Package tags manages Tags on Compute V2 servers.

This extension is available since 2.26 Compute V2 API microversion.

Example to List all server Tags

	client.Microversion = "2.26"

    serverTags, err := tags.List(client, serverID).Extract()
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Tags: %v\n", serverTags)

Example to Check if the specific Tag exists on a server

    client.Microversion = "2.26"

    exists, err := tags.Check(client, serverID, tag).Extract()
    if err != nil {
        log.Fatal(err)
    }

    if exists {
        log.Printf("Tag %s is set\n", tag)
    } else {
        log.Printf("Tag %s is not set\n", tag)
    }

Example to Replace all Tags on a server

    client.Microversion = "2.26"

    newTags, err := tags.ReplaceAll(client, serverID, tags.ReplaceAllOpts{Tags: []string{"foo", "bar"}}).Extract()
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("New tags: %v\n", newTags)

Example to Add a new Tag on a server

    client.Microversion = "2.26"

    err := tags.Add(client, serverID, "foo").ExtractErr()
    if err != nil {
        log.Fatal(err)
    }

Example to Delete a Tag on a server

    client.Microversion = "2.26"

    err := tags.Delete(client, serverID, "foo").ExtractErr()
    if err != nil {
        log.Fatal(err)
    }

Example to Delete all Tags on a server

    client.Microversion = "2.26"

    err := tags.DeleteAll(client, serverID).ExtractErr()
    if err != nil {
        log.Fatal(err)
    }
*/
package tags
