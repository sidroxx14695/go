I am calling parseGraphqlQuery method which will have some common checks for header empty or not, error in reading request body etc
I generate the fieldmap
Taking in multiple policy keys and split them with comma
unmarshalling the json requests body 
collectKeys method is to traverse the nested json structure, extracting and returning the full path of fields in a dotted notation. This is to allow effectively locating and removing specific fields within json based on their path.
——explanation for collectKeys———
calling processJsonData that will perform the redaction for each policy key using iteration
this method takes json request, fieldmap, policy and collectKeys result
if the policy key has Query as typename in policy key then we will perform deletion at root level( same code which we reviewed last time)
if dataField, ok := jsonStr["data"].(map[string]any); ok {  -> It asserts that “data” key in the jsonStr is of type map[string]any (map of keys of type string and the value of nay type)
deleteKey hold the field name against the typename that matches the type name of the policy key provided. Example if the policy key is AvailableCreditAmount.availableSpendingCreditAmount then deleteKey will hold availableCreditAmount
After the dotted key path is formed we need to perform deletion so we iterate over each key in collectKeys and find the last and second last keys. for example -> for our policy key AvailableCreditAmount.availableSpendingCreditAmount,  in our collectKeys we have an entry data.getAccount.availableCreditAmount.availableSpendingCreditAmount so here the lastkey is availableSpendingCreditAmount and second lastkey is availableCreditAmount so now I will match if the deletekey that has the availableCreditAmount matches with secondLastKey and the policy[1] that has availableSpendingCreditAmount matches with lastkey if so then I will delete that path from the json
removeArrayIndices will remove the indexes mentioned for some path which has array type structure like data.getAccount.availableCreditAmount.availableCashCreditAmount data.getAccount.cards data.getAccount.cards[0].cardNumber 
deleteJSONPath method will delete the key based on path sent to it.
         parentPath := path[:len(path)-len(path[strings.LastIndex(path, ".")+1:])-1] -> Finds the last occurence of dot in the path  and gets slice of path up to the last dot, excluding the key to be deleted
Lookup will find the parent structure in the json
Finally find the key to be deleted from the path and delete that from the json


collectKeys explanation

1. *Initialization*: It starts with an empty list to store the full paths of keys.

2. *Traversal*: It goes through each key-value pair in the JSON object.

3. *Path Construction*: For each key, it builds a "path" that represents its location in the JSON object. If it's a nested object, it adds the current key to the path and continues into the nested object.

4. *Recursion*: If a value is another JSON object (a nested map), it recursively processes that object, building paths for its keys as well.

5. *Collection*: It collects all these paths into a list, representing the complete paths to each key in the JSON structure.

6. *Return*: Finally, it returns the list of paths, which can be used to find or delete specific fields in the JSON.

This process effectively flattens the JSON structure into a list of key paths, making it easier to manage specific fields.


Complexity:

deleteJSONPath:

Finding the key → O(N) (depends on depth of JSON).
Deleting it → O(1)
Overall: O(N)

collectKeys:

Scanning all fields in json  → O(N)
Handling deep JSON → O(D) (depends on depth).
Overall: O(N + D) (Still Efficient, but recursion can be slow).

Happens one time.

processJsonData:

Checking each key against policy → O(K) (looping through json paths).
Deleting keys → O(N) —-> which is for deleteJSONPath
Overall: O(N + K) (Can be Slow for Big JSONs).

ParseGraphQLQuery:

collectKeys -> O(N + D)
processJsonData is called for each policy so overall will get complexity of : O((N + K) times P) where P is number of policies.

Overall complexity: O((N+D)+((N + K) times P))
‘


——————————————————————————————————-


A GraphQL JSON data payload (parsed as map[string]any)
A list of policies in the form: TypeName.FieldName (e.g., AvailableCreditAmount.availableCashCreditAmount)
A field map (from your schema) that maps each field path to its corresponding type name

⚙️ Step-by-Step Algorithm
ProcessJsonData is method being called where earlier we had just one policy key passed to it but now we have multiple policy keys and using that I am forming policyMap
policyMap := make(map[string]map[string]bool) 
stores the typename and filename of the policy keys as a set
This simulates a set in Go since Go doesn’t have a built-in set type. Using a map with bool values is the most common workaround. I am using set since traversal with set is O(1)
🔁 1. Parse and Preprocess
Read the Policies header, split it by commas.
For each policy:
If it's Query.XYZ, remove the top-level field XYZ from data["data"]
Otherwise, resolve the actual parent key name (e.g., AvailableCreditAmount) using allFieldMap
Store each TypeName.FieldName as a lookup entry in policyMap:

policyMap[parentKey][fieldName] = true

🌳 2. Traverse and Delete – collectKeysAndDeleteMulti
Recursively visit every key in the JSON (depth-first traversal).
For each field:
Track its full dotted path: e.g., data.profile.availableCreditAmount.availableCashCreditAmount
Extract the last two segments (field and its parent object)
Strip array indices if present ([0], [1] etc.)

Example:
lastKey = availableCashCreditAmount
secondLastKey = availableCreditAmount

if policyMap[secondLastKey][lastKey] exists then
    delete current key from map

Recurse into nested objects and arrays

💡 Why this is optimal
✅ Single traversal of the JSON✅ Constant-time lookup for deletions✅ No need to compute paths again or traverse more than once✅ Works for nested fields and arrays




Prabu I have made changes so that deletion happens with just single traversal of the json. I will go over the updates quickly
🔧 processJsonData(…)
ProcessJsonData is method being called where earlier we had just one policy key passed to it but now we have multiple policy keys and using that I am forming policyMap

🔍 What it does:
Validates input: Makes sure JSON and policies are present.
Creates policyMap:policyMap := map[typename][fieldname]bool
To process each policy key (like AvailableCreditAmount.availableCashCreditAmount),it builds a map of fields i.e. policyMap that need to be deleted
policyMap stores the typename and filename of the policy keys as a set
This simulates a set in Go since Go doesn’t have a built-in set type. Using a map with bool values is the most common workaround. I am using set since traversal with set is O(1)


Example:{
  "availableCreditAmount": {
    "availableCashCreditAmount": true,
    "availableRewardCreditAmount": true
  },
  "account": {
    "accountReferenceId": true
  }
}

It acts like a lookup set to quickly check if a field needs deletion.
Handles root-level fields: If the policy is like Query.accounts, it directly deletes accounts from the top-level "data" field.
Finds the key (from schema map) that matches the typename and adds its field(s) to the policyMap.
Calls collectKeysAndDelete once to do the actual deletion ✂️

🔁 collectKeysAndDelete(...)
Purpose:
This is a recursive JSON walker. It traverses your full nested JSON tree once and deletes any key that matches a (typename, fieldname) from policyMap.
🔍 What it does:
For every key in the map, it constructs the full dotted path:
E.g., data.profile.availableCreditAmount.availableCashCreditAmount
Strips array indices so it works even for lists (like items[0].amount)
Extracts last and second last segments from the dotted path:
E.g., from ...availableCreditAmount.availableCashCreditAmount
lastKey = availableCashCreditAmount
secondLastKey = availableCreditAmount
Checks:if policyMap[secondLastKey][lastKey] == true
If true → delete that key from the map.

Recurses into:
Nested objects (maps)
Arrays of objects


So Now we do,
✅ Single traversal of the JSON✅ have Constant-time lookup for deletions✅ No need to compute paths again or traverse more than once✅ it Works for nested fields and arrays



As discussed last time I have modified the method where redaction is happening
traverseAndRedact is that method
it is taking parameters :
json which is formed after unmarshalling the request input json
fieldMap - map of all fields and their type
policyMap - has all the policies to be deleted in form of a set. This a map of maps where key is typeName and value is a map and this  furthur contains the fieldname as key …and value as a boolean which is always true and this is how set is implemented in Go since there is no in built set collection. But since it is a set in Go …there is no need to iterate over multiple policy keys if passed in header, with set the traversal will be just O(1) when looking for a given policy key unlike O(n) when doing same with say an array of policy keys.
typename - is empty at first and will hold the type for the field being processed currently when traversing the json

what this method does:
It will iterate over each key and value pair of json 
It checks if the fieldType for the currentKey is present in fieldMap. If so, it updates the typeName with the typename that is present in fieldMap
Normalize method is used in case we have typeName with square brackets [ ], then we need to remove those brackets. For example:       the type for accounts is array of Account in fieldmap("accounts": "[Account]") so when we have policy key as Account.cards so when redacting we need to take into consideration all the arrays which is of type Account along with the fields whose type is Account directly… like getAccount.   so for that we need to remove the square brackets [ ] from the typeName and then compare that with the type in policyMap and continue with redaction.
If the typeName matches a key in policyMap and the current fieldName matches the value then we will delete the fieldname directly from the json.
and recusively we go through the entire json where we check the type for each field in switch statement.                                               
                 1.  If the type of value is a map of string and any type then we call the traverseAndRedact again
               
                 2. If the type of value is array then we call the traverseAndRedact again for each object in that array





In code what I am doing now is I have included the PolicyMap to have the engineResponse also added to it for each PolicyKey

getEngineResponseBasedOnPolicy
1. We are calling method called getEngineResponseBasedOnPolicy where pass the policy key and get the engine response
2. If there is no engine response found for a policy key then we return null

And then we add the engineResponse fetched to the policyMap for that policyKey

And the calling the method redactPolicyKeys
1. redactPolicyKeys method is now having new RefId parameter added which will hold the reference id for a given context so for Account it will be account reference id and for card it will be card reference id
2. At start it is empty at start
3. List earlier we go through all the fields in the json but now if we are in Account context by checking the typename then we find the accountReferenceId and store it in refId if found
4. Same for Card , if in Card context then we find the cardRerencenceId in the context and store it in refId if found
5. And now if the Typename and fieldName matched the one present in policyMap then go and check the new method called processEngineResponse 
6. Where we pass the engineResponse for the given policyKey as stored in PolicyMap
7. In processEngineResonse method we pass the engineResponse and the referenceID
8. In engineResponse we look for the referenceID that matches the referenceId of that context we are in and then return true if the value is ALLOW or false if value is TRUE
9. If the processEngineResonse returns false then we delete that field
10. And then same as before we recursively call the redactPolicyKeys method for all objects and array type fields.

I am yet to test this with alll different policy key types but currently it is working for cases like Account.balance, Card.cardNumber and even AvailableCreditAmount.availableSpendingCreditAmount

This is my progress till ow



