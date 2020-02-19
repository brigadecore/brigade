/*
Package resourcetypes provides operations for listing available resource types,
obtaining their properties schema, and generating example templates that can be
customised to use as provider templates.

Example of listing available resource types:

    listOpts := resourcetypes.ListOpts{
        SupportStatus: resourcetypes.SupportStatusSupported,
    }

    resourceTypes, err := resourcetypes.List(client, listOpts).Extract()
    if err != nil {
        panic(err)
    }
    fmt.Println("Get Resource Type List")
    for _, rt := range resTypes {
        fmt.Println(rt.ResourceType)
    }
*/
package resourcetypes
