pub mod crdt {
    use uuid::Uuid;
    use std::collections::HashSet;
    use std::collections::HashMap;
    /*
        To help: PN-Counter for Item of a list

        payload integer[n] P, integer[n] N
            initial [0,0,...,0], [0,0,...,0]
        update increment()
            let g = myId()
            P[g] := P[g] + 1
        update decrement()
            let g = myId()
            N[g] := N[g] + 1
        query value() : integer v
            let v = Σi P[i] - Σi N[i]
        compare (X, Y) : boolean b
            let b = (∀i ∈ [0, n - 1] : X.P[i] ≤ Y.P[i] ∧ ∀i ∈ [0, n - 1] : X.N[i] ≤ Y.N[i])
        merge (X, Y) : payload Z
            let ∀i ∈ [0, n - 1] : Z.P[i] = max(X.P[i], Y.P[i])
            let ∀i ∈ [0, n - 1] : Z.N[i] = max(X.N[i], Y.N[i]
        
    */
    // Only grow Counter for context

    


    
    //TODO: check if we can use this: This CRDT is restricted to increment/decrement directly the amount we want and input can only be positive
    #[derive(Clone,Debug)]
    pub struct BoundedPNCounterv2 {
            pub positive_count: HashMap<Uuid, u32> ,
            pub negative_count: HashMap<Uuid, u32>,
        }

    impl BoundedPNCounterv2{
        pub fn new() -> Self {
                BoundedPNCounterv2 {
                    positive_count: HashMap::new(),
                    negative_count: HashMap::new(),
                }
            }

        pub fn positive_count(&self) -> &HashMap<Uuid, u32> {
            &self.positive_count
        }
    
        pub fn negative_count(&self) -> &HashMap<Uuid, u32> {
            &self.negative_count
        }
        pub fn increment(&mut self, node_id: Uuid, amount: u32) {
            let inc_count = self.positive_count.entry(node_id).or_insert(0);
            //deal with overflow-> checked on property tests!
            if let Some(new_value) = inc_count.checked_add(amount) {
                *inc_count = new_value;
            }
            // If overflow occurs, the operation is ignored
             
        }
    
        pub fn decrement(&mut self, node_id: Uuid, amount: u32) {
            let dec_count = self.negative_count.entry(node_id).or_insert(0);
            let positive_count = self.positive_count.entry(node_id).or_insert(0);
            if let Some(new_value) = dec_count.checked_add(amount) {
                if new_value > *positive_count {
                    *dec_count = *positive_count;  
                } else {
                    *dec_count =new_value;
                }
                
            }
            // if overflow, ignore
        }
    

        pub fn get_count(&self) -> u32 {
            let sum_pos_count: u32 = self.positive_count.values().sum();
            let sum_neg_count: u32 = self.negative_count.values().sum();

            sum_pos_count - sum_neg_count
        
        }

        pub fn compare(&self, inc_b_pn_counter: &BoundedPNCounterv2) -> bool{
            for node_id in self.positive_count().keys() {
                let pos_val1 = self.positive_count().get(node_id).unwrap_or(&0);
                let pos_val2 = inc_b_pn_counter.positive_count().get(node_id).unwrap_or(&0);

                if pos_val1 > pos_val2 {
                    return false;
                }
            }

            for node_id in self.negative_count().keys() {
                let neg_val1 = self.negative_count().get(node_id).unwrap_or(&0);
                let neg_val2 = inc_b_pn_counter.negative_count().get(node_id).unwrap_or(&0);

                if neg_val1 > neg_val2 {
                    return false;
                }
            }

            true

        }
        // merge function perserving: commutative, associative, and idempotent.
        pub fn merge(&self, other: &BoundedPNCounterv2) -> BoundedPNCounterv2 {
            let mut merged = BoundedPNCounterv2::new();
    
            // Merge positive counts
            let all_keys = self.positive_count.keys().chain(other.positive_count.keys()); // Here We need to get all possible keys in either counters
            for &node_id in all_keys {
                let self_count = self.positive_count.get(&node_id).unwrap_or(&0);
                let other_count = other.positive_count.get(&node_id).unwrap_or(&0);
                merged.positive_count.insert(node_id, std::cmp::max(*self_count, *other_count));
            }

            // Merge negative counts
            let all_keys = self.negative_count.keys().chain(other.negative_count.keys()); // same as above
            for &node_id in all_keys {
                let self_count = self.negative_count.get(&node_id).unwrap_or(&0);
                let other_count = other.negative_count.get(&node_id).unwrap_or(&0);
                merged.negative_count.insert(node_id, std::cmp::max(*self_count, *other_count));
            }

            merged
        }
    }

    

    // // Arranjar estratégias de compressão para os states dos CRDTs !!! Passamos o estado, com o tempo isto vai acumular muita informação


    // Formulation:
    // List of (Items(Name,quanity: PN-Counter), (Nodeid/clientId, only grow counter)
    //TODO: Use MAP, so we just update counters, we dont need to save every 3-tuple, 2-tuple
    #[derive(Clone, Debug)]
    pub struct AWSet {
        pub state: HashSet<(String, Uuid, u32)>, // Set of tuples (Item, NodeId, Counter)
        pub context: HashSet<(Uuid, u32)>, // Set of tuples (NodeId, Counter)
    }
    impl AWSet {
        pub fn new() -> Self {
            //TODO: check if Hashset is the best struture to use, do we need to save every 3-tuple and 2-tuple on state and context, or just the more recent one
            AWSet{
                state: HashSet::new(),
                context: HashSet::new(), 
            }
        }
        

        // get elements ( Items) of AWSet with corresponding state and context
        pub fn elements(&self) -> Vec<String> {
            let mut unique_items = HashSet::new();
                for (item_name, _, _) in &self.state {
                    unique_items.insert(item_name.clone());
                }
                unique_items.into_iter().collect()
        }
        // Check if given element (Item) with corresponding state and context, exist on AWSet
        pub fn contains(&self, item_name: &str) -> bool {
            self.state.iter().any(|(name, _, _)| name == item_name)
        }

        
        
        pub fn max_i(&self, node_id: Uuid) -> u32 {
            self.context.iter()
                .filter(|(uuid, _)| *uuid == node_id)
                .map(|(_, counter)| *counter)
                .max()
                .unwrap_or(0)//if there is no (NodeId, Counter) tuple, max_i returns 0, so nex_i can generate tuples(node_id, 0+1( Counter)), when there is no pair on the context
        }
    
        pub fn next_i(&self, node_id: Uuid) -> (Uuid, u32) {
            (node_id, self.max_i(node_id) + 1) 
        }
         //this add do a "type" of garbage colection: we just mantain the updated context/state tuples
        pub fn add_i(&mut self, item_name: String, node_id: Uuid) {
            let new_counter = self.next_i(node_id).1;

            

            // We remove the old tuple (node_id, counter) if it exists
            self.context.retain(|(id, _)| *id != node_id);

            // Context is updated
            self.context.insert((node_id, new_counter)); // c ∪ {d}
            

            // Remove the old tuple (element, node_id, counter) if it exists
            self.state.retain(|(name, id, _)| !(name == &item_name && id == &node_id)); // s ∪ {(e, d)} -> this version removes old (e,d')

            //State is updated
            self.state.insert((item_name, node_id, new_counter));

        }

        //Here we remove all tuples from state, that have the item_name
        pub fn rmv_i(&mut self, item_name: String) {
            self.state.retain(|(name,_,_)| *name != item_name);
            
        }

        pub fn filter(&self, inc_awset: &AWSet) -> HashSet<(String, Uuid, u32)> {
            self.state.iter()
            .filter(|(_name, node_id, counter)| {
                !inc_awset.context.iter().any(|(inc_node_id,inc_counter)|{
                    node_id == inc_node_id && counter < inc_counter 
                })
            })
            .cloned()
            .collect()
        }

        pub fn merge(&mut self, inc_awset: &AWSet){
            //Intersection between states of two AWSets
            let states_intersection = self.state.intersection(&inc_awset.state).cloned().collect();
            // Union of filter(s,c') U f(s',c)
            let filter_state_1: HashSet<_> = self.filter(&inc_awset);
            let filter_state_2: HashSet<_> = inc_awset.filter(&self);

            let union_12: HashSet<_> = filter_state_1.union(&filter_state_2).cloned().collect();
            let final_merge: HashSet<_> = union_12.union(&states_intersection).cloned().collect();

            
            // Union of contexts
            let final_context: HashSet<_> = self.context.union(&inc_awset.context).cloned().collect();
            
            // ending merge
            self.state = final_merge;
            self.context = final_context;
        }
    }
    
    
    
    #[derive(Clone, Debug)]
    pub struct ShoppingList{
        pub node_id : Uuid,
        pub items: HashMap<String, BoundedPNCounterv2 >,
        pub awset: AWSet,
    }
    impl ShoppingList{
        pub fn new() -> Self{
            ShoppingList{
                node_id: Uuid::new_v4(), // node to identify the user in a distributed system
                items: HashMap::new(), 
                awset: AWSet::new(),}
        }
        //Add-Wins
        pub fn add_or_update_item(&mut self, item_name: String, quantity_change: i32){
            self.awset.add_i(item_name.clone(), self.node_id);
            

            
            //using the information in awset we add_or_update the Map of Items
            if let Some(existing_item) = self.items.get_mut(&item_name) {
                    
                    if quantity_change < 0 {
                        let dec_quant_change: u32 = (-1 * quantity_change).try_into().unwrap();
                        existing_item.decrement(self.node_id, dec_quant_change.try_into().unwrap());

                    }else if quantity_change > 0{
                        existing_item.increment(self.node_id, quantity_change.try_into().unwrap());
                    }
                    

            }
            else{ //Item doesn't exist on Map of Items
                
                let mut new_item = BoundedPNCounterv2::new();
                if quantity_change < 0 {
                    let dec_quant_change = -1 * quantity_change;
                    new_item.decrement(self.node_id, dec_quant_change.try_into().unwrap());
                    self.items.insert(item_name, new_item);

                    

                }else if quantity_change >= 0{
                    new_item.increment(self.node_id, quantity_change.try_into().unwrap());
                    self.items.insert(item_name, new_item);
                }
            }

                
        }

        pub fn remove_item(&mut self, item_name: String){
            self.awset.rmv_i(item_name.clone());
            self.items.remove(&item_name);
        }

        
        pub fn merge(&mut self, inc_list: &ShoppingList) {
            
            self.awset.merge(&inc_list.awset);
    
            
            let mut merged_items = HashMap::new();

            let elements = self.awset.elements();

            // Merge items based on the merged AWSet
            for item_name in self.elements {
                let merged_item = match (self.items.get(&item_name), inc_list.items.get(&item_name)) {
                    (Some(self_item), Some(inc_list_item)) => self_item.merge(inc_list_item),
                    (None, Some(inc_list_item)) | (Some(inc_list_item), None) => inc_list_item.clone(),
                    _ => continue, // Skip items not present in either list
                };

                merged_items.insert(item_name, merged_item);
            }

            // Update items with merged results
            self.items = merged_items;
        }


        // Get all items names
        pub fn get_items(&self) -> Vec<String> {
            self.awset.elements()
        }
    }

    
    
}
    
#[cfg(test)]
pub mod tests {
    use crate::crdt::crdt::*;
    use uuid::Uuid;
    use std::collections::HashSet;

    

    //TODO: //Unit Tests for Bounded_PNCounterv2 ( with amount) to increment
    #[test]
    fn test_increment_bounded_pncounter() {
        let mut counter = BoundedPNCounterv2::new();
        let node_id = Uuid::new_v4();
        counter.increment(node_id, 5);
        assert_eq!(*counter.positive_count.get(&node_id).unwrap(), 5);
    }

    #[test]
    fn test_decrement_bounded_pncounter() {
        let mut counter = BoundedPNCounterv2::new();
        let node_id = Uuid::new_v4();
    
        
        counter.increment(node_id, 5);
    
        
        counter.decrement(node_id, 3);
    
        
        assert_eq!(*counter.negative_count.get(&node_id).unwrap(), 3);
    }

    #[test]
    fn test_get_count_bounded_pncounter() {
        let mut counter = BoundedPNCounterv2::new();
        let node_id = Uuid::new_v4();
        counter.increment(node_id, 10);
        counter.decrement(node_id, 4);
        assert_eq!(counter.get_count(), 6);
    }

    #[test]
    fn test_lower_boundary_bounded_pncounter() {
        let mut counter = BoundedPNCounterv2::new();
        let node_id = Uuid::new_v4();
        counter.increment(node_id, 10);
        counter.decrement(node_id, 4);
        counter.decrement(node_id, 10);  // Assuming decrement is bounded by positive count
        assert_eq!(counter.get_count(), 0);
    }

    #[test]
    fn test_bounded_pncounter_compare() {
        let mut counter1 = BoundedPNCounterv2::new();
        let mut counter2 = BoundedPNCounterv2::new();
        let node_id = Uuid::new_v4();

        counter1.increment(node_id, 3); // Decrease to 3, making it less than or equal to counter2
        counter2.increment(node_id, 5);

        assert!(counter1.compare(&counter2)); // Now this should pass
    }
    #[test]
    fn test_merge_same_keys() {
        let mut counter1 = BoundedPNCounterv2::new();
        let mut counter2 = BoundedPNCounterv2::new();

        let node_id = Uuid::new_v4();
        counter1.increment(node_id, 2);
        counter2.increment(node_id, 3);

        let merged = counter1.merge(&counter2);
        assert_eq!(*merged.positive_count.get(&node_id).unwrap(), 3);
    }

    #[test]
    fn test_merge_disjoint_keys() {
        let mut counter1 = BoundedPNCounterv2::new();
        let mut counter2 = BoundedPNCounterv2::new();

        let node1 = Uuid::new_v4();
        let node2 = Uuid::new_v4();
        counter1.increment(node1, 2);
        counter2.increment(node2, 3);

        let merged = counter1.merge(&counter2);
        assert_eq!(*merged.positive_count.get(&node1).unwrap(), 2);
        assert_eq!(*merged.positive_count.get(&node2).unwrap(), 3);
    }

    #[test]
    fn test_merge_empty_counters() {
        let counter1 = BoundedPNCounterv2::new();
        let counter2 = BoundedPNCounterv2::new();

        let merged = counter1.merge(&counter2);
        assert!(merged.positive_count.is_empty());
        assert!(merged.negative_count.is_empty());
    }

    #[test]
    fn test_merge_one_empty_counter() {
        let mut counter1 = BoundedPNCounterv2::new();
        let counter2 = BoundedPNCounterv2::new();

        let node_id = Uuid::new_v4();
        counter1.increment(node_id, 1);

        let merged = counter1.merge(&counter2);
        assert_eq!(*merged.positive_count.get(&node_id).unwrap(), 1);
    }

    
    // //Test AWSet

   

    #[test]
    fn test_awset_new() {
        let awset = AWSet::new();
        assert!(awset.state.is_empty());
        assert!(awset.context.is_empty());
    }

    #[test]
    fn test_max_i() {
        let mut awset = AWSet::new();
        let node_id = Uuid::new_v4();
        awset.context.insert((node_id, 1));
        awset.context.insert((node_id, 3));
        awset.context.insert((node_id, 2));

        assert_eq!(awset.max_i(node_id), 3);
    }

    #[test]
    fn test_next_i() {
        let mut awset = AWSet::new();
        let node_id = Uuid::new_v4();
        awset.context.insert((node_id, 1));
        awset.context.insert((node_id, 2));

        let next = awset.next_i(node_id);
        assert_eq!(next, (node_id, 3));
    }


    #[test]
    fn test_context_with_multiple_nodes() {
        let mut awset = AWSet::new();
        let node_id1 = Uuid::new_v4();
        let node_id2 = Uuid::new_v4();

        
        awset.context.insert((node_id1, 1));
        awset.context.insert((node_id1, 2));
        awset.context.insert((node_id2, 1));
        awset.context.insert((node_id2, 3));

        // Test max_i for different nodes
        assert_eq!(awset.max_i(node_id1), 2);
        assert_eq!(awset.max_i(node_id2), 3);

        // Test next_i for different nodes
        let next1 = awset.next_i(node_id1);
        let next2 = awset.next_i(node_id2);
        assert_eq!(next1, (node_id1, 3));
        assert_eq!(next2, (node_id2, 4));
    }

    #[test]
    fn test_add_new_item() {
        let mut awset = AWSet::new();
        let node_id = Uuid::new_v4();
        let item_name = "apple".to_string();
        

        awset.add_i(item_name.clone(), node_id);
        
       
        assert!(awset.state.contains(&(item_name, node_id, 1)));
        assert!(awset.context.contains(&(node_id, 1)));
    }

    #[test]
    fn test_increment_existing_item() {
        let mut awset = AWSet::new();
        let node_id = Uuid::new_v4();
        let item_name = "apple".to_string();
        awset.state.insert((item_name.clone(), node_id, 1));
        awset.context.insert((node_id, 1));

        awset.add_i(item_name.clone(), node_id);
        
        
        assert!(awset.state.contains(&(item_name, node_id, 2)));
        assert!(awset.context.contains(&(node_id, 2)));
    }

    #[test]
    fn test_decrement_existing_item() {
        let mut awset = AWSet::new();
        let node_id = Uuid::new_v4();
        let item_name = "apple".to_string();
        awset.state.insert((item_name.clone(), node_id, 1));
        awset.context.insert((node_id, 1));

        awset.add_i(item_name.clone(), node_id);
        
        
        assert!(awset.state.contains(&(item_name, node_id, 2)));
        assert!(awset.context.contains(&(node_id, 2)));
    }

    fn test_add_i() {
        let mut awset = AWSet::new();
        let node_id = Uuid::new_v4();
        let item_name = "apple".to_string();

       
        awset.add_i(item_name.clone(), node_id);
        assert_eq!(awset.state.contains(&(item_name.clone(), node_id, 1)), true);
    }

    // Unit tests for rmv_i
    #[test]
    fn test_rmv_i_existing_item() {
        let mut awset = AWSet::new();
        let node_id = Uuid::new_v4();
        let item_name = "apple".to_string();

        
        awset.add_i(item_name.clone(), node_id);
        awset.rmv_i(item_name.clone());

        assert_eq!(awset.state.contains(&(item_name, node_id, 1)), false);
    }

    
    #[test]
    fn test_rmv_i_non_existent_item() {
        let mut awset = AWSet::new();
        let node_id = Uuid::new_v4();
        let item_name = "apple".to_string();
        let non_existent_item = "banana".to_string();

        
        awset.add_i(item_name.clone(), node_id);
        awset.rmv_i(non_existent_item);

        // Original item should still exist
        assert_eq!(awset.state.contains(&(item_name, node_id, 1)), true);
    }

    
    #[test]
    fn test_rmv_i_context_unchanged() {
        let mut awset = AWSet::new();
        let node_id = Uuid::new_v4();
        let item_name = "apple".to_string();

        
        awset.add_i(item_name.clone(), node_id);
        let context_before_removal = awset.context.clone();
        awset.rmv_i(item_name);

        // Context should remain the same
        assert_eq!(awset.context, context_before_removal);
    }

    #[test]
    fn test_filter_function() {
        let node_id_1 = Uuid::new_v4();
        let node_id_2 = Uuid::new_v4();
        let mut awset_1 = AWSet::new();
        let mut awset_2 = AWSet::new();

        
        awset_1.state.insert(("apple".to_string(), node_id_1, 1));
        awset_1.context.insert((node_id_1, 2));

        awset_2.state.insert(("banana".to_string(), node_id_2, 1));
        awset_2.context.insert((node_id_2, 2));

        // Expected result after filtering awset_1 against awset_2: Mock
        let mut expected_state: HashSet<(String, Uuid, u32)> = HashSet::new(); //"apple" should not be in the filtered state
        expected_state.insert(("apple".to_string(), node_id_1, 1));
        
        let filtered_state = awset_1.filter(&awset_2);

        // Check that the filtered state matches the expected state
        assert_eq!(filtered_state, expected_state);
    }

    //Test merge

    #[test]
    fn test_merge_with_overlap() {
        let mut awset1 = AWSet::new();
        let mut awset2 = AWSet::new();

        let node_id1 = Uuid::new_v4();
        let node_id2 = Uuid::new_v4();
        let counter1 = 1;
        let counter2 = 2;

        
        awset1.state.insert(("apple".to_string(), node_id1, counter1));
        awset2.state.insert(("apple".to_string(), node_id2, counter2));

        // Merging should result in a set that contains both items
        awset1.merge(&awset2);
        assert_eq!(awset1.state.len(), 2);
    }

    #[test]
    fn test_merge_with_distinct_items() {
        let mut awset1 = AWSet::new();
        let mut awset2 = AWSet::new();

        let node_id1 = Uuid::new_v4();
        let node_id2 = Uuid::new_v4();
        let counter1 = 1;
        let counter2 = 2;

       
        awset1.state.insert(("apple".to_string(), node_id1, counter1));
        awset2.state.insert(("banana".to_string(), node_id2, counter2));

        // Merging should result in a set that contains both items
        awset1.merge(&awset2);
        assert_eq!(awset1.state.len(), 2);
    }

    #[test]
    fn test_merge_with_unique_items() {
        let mut awset1 = AWSet::new();
        let awset2 = AWSet::new();

        let node_id = Uuid::new_v4();
        let counter = 1;

        
        awset1.state.insert(("apple".to_string(), node_id, counter));

        // Merging with an empty set should not change the first set
        awset1.merge(&awset2);
        assert_eq!(awset1.state.len(), 1);
        assert!(awset1.state.contains(&("apple".to_string(), node_id, counter)));
    }

    #[test]
    fn test_merge_with_empty_sets() {
        let mut awset1 = AWSet::new();
        let awset2 = AWSet::new();

        // Merging two empty sets should result in an empty set
        awset1.merge(&awset2);
        assert!(awset1.state.is_empty());
    }

    #[test]
    fn test_elements() {
        let mut awset = AWSet::new();
        let node_id = Uuid::new_v4();

        // Adding items
        awset.add_i("Apple".to_string(), node_id);
        awset.add_i("Banana".to_string(), node_id);
        awset.add_i("Apple".to_string(), node_id); // Add on existing item

        let elements = awset.elements();
        assert_eq!(elements.len(), 2); // Should contain 2 unique items
        assert!(elements.contains(&"Apple".to_string()));
        assert!(elements.contains(&"Banana".to_string()));
    }

    #[test]
    fn test_contains() {
        let mut awset = AWSet::new();
        let node_id = Uuid::new_v4();

        // Adding items
        awset.add_i("Apple".to_string(), node_id);

        assert!(awset.contains("Apple"));
        assert!(!awset.contains("Banana"));
    }

    
}

#[cfg(test)]
mod integration_tests {
    // use crate::crdt::crdt::*;
    // use uuid::Uuid;


    // #[test]
    // fn test_add_merge() {
    //     let mut awset1 = AWSet::new();
    //     let mut awset2 = AWSet::new();
    //     let node_id1 = Uuid::new_v4();
    //     let node_id2 = Uuid::new_v4();

        
    //     awset1.add_i("apple".to_string(), node_id1);
    //     awset2.add_i("banana".to_string(), node_id2);

        
    //     awset1.merge(&awset2);

    //    // Both id's have added, so the corresponding states are the above
    //     assert!(awset1.state.contains(&("apple".to_string(), node_id1, 1)));
    //     assert!(awset1.state.contains(&("banana".to_string(), node_id2, 1)));
    //     assert!(awset1.context.contains(&( node_id1, 1)));
    //     assert!(awset1.context.contains(&( node_id2, 1)));
    // }

    // #[test]
    // fn test_remove_merge() {
    //     let mut awset1 = AWSet::new();
    //     let mut awset2 = AWSet::new();
    //     let node_id1 = Uuid::new_v4();
    //     let node_id2 = Uuid::new_v4();

        
    //     awset1.add_i("apple".to_string(), node_id1);
    //     awset1.add_i("apple".to_string(), node_id1);
    //     awset2.add_i("apple".to_string(), node_id2);
    //     awset2.rmv_i("apple".to_string());

     
    //     awset1.merge(&awset2);

    //     //In the merge, remove needs to stay because id1 have a greater causal context ( Counter =2)
    //     //So Add wins policy is used
    //     assert!(awset1.state.contains(&("apple".to_string(), node_id1, 2)));
    //     assert!(awset1.context.contains(&( node_id1, 2)));
    //     assert!(!awset2.state.contains(&("apple".to_string(), node_id2, 1)));
    //     assert!(awset2.context.contains(&(node_id2, 1)));
    // }

    // #[test]
    // fn test_add_remove_same_item_merge() {
    //     let mut awset1 = AWSet::new();
    //     let mut awset2 = AWSet::new();
    //     let node_id1 = Uuid::new_v4();
    //     let node_id2 = Uuid::new_v4();

       
    //     awset1.add_i("apple".to_string(), node_id1);
    //     awset1.rmv_i("apple".to_string());
    //     awset2.add_i("apple".to_string(), node_id2);
    //     awset2.rmv_i("apple".to_string());

       
    //     awset1.merge(&awset2);
    //     //Both concurrently want to remove apple, remove is done
    //     assert!(!awset1.state.contains(&("apple".to_string(), node_id1, 1)));
    //     assert!(!awset1.state.contains(&("apple".to_string(), node_id2, 1)));
    //     assert!(awset1.context.contains(&( node_id1, 1)));
    //     assert!(awset2.context.contains(&( node_id2, 1)));
    // }


    // #[test]
    // fn test_remove_multiple_add_merge() {
    //     let mut awset1 = AWSet::new();
    //     let mut awset2 = AWSet::new();
    //     let node_id1 = Uuid::new_v4();
    //     let node_id2 = Uuid::new_v4();

        
    //     awset1.add_i("apple".to_string(), node_id1);
    //     awset1.add_i("orange".to_string(), node_id1);
    //     awset1.rmv_i("apple".to_string());
    //     awset2.add_i("apple".to_string(), node_id2);
        

     
    //     awset1.merge(&awset2);

    //     // Apple needs to be removed, nodes id1/id2 have the same causal context (Counter =2)
    //     // It's like we only have one remove, without concurrent add
    //     assert!(awset1.state.contains(&("apple".to_string(), node_id2, 1))); // awset1 after merge will continue have item apple, concurrent add from awset2 wins!
    //     assert!(awset1.context.contains(&( node_id1, 2)));
    //     assert!(awset1.state.contains(&("orange".to_string(), node_id1, 2)));
        
    //     assert!(awset2.state.contains(&("apple".to_string(), node_id2, 1)));
    //     assert!(awset2.context.contains(&(node_id2, 1)));
        
    // }

    // #[test]
    // fn test_remove_multiple_add_merge2() {
    //     let mut awset1 = AWSet::new();
    //     let mut awset2 = AWSet::new();
    //     let node_id1 = Uuid::new_v4();
    //     let node_id2 = Uuid::new_v4();

        
    //     awset1.add_i("apple".to_string(), node_id1);
    //     awset1.add_i("apple".to_string(), node_id1);
    //     awset2.add_i("apple".to_string(), node_id2);
    //     awset2.add_i("apple".to_string(), node_id2);
    //     awset2.add_i("apple".to_string(), node_id2);
    //     awset2.rmv_i("apple".to_string());

     
    //     awset1.merge(&awset2);

    //     // Apple needs to be removed, nodes id2 have a greater causal context (Counter =2)
    //     // It's like id2, is the last user to add apples, and now he wants to remove without concurrents adds from other users
    //     assert!(awset1.state.contains(&("apple".to_string(), node_id1, 2)));
    //     assert!(awset1.context.contains(&( node_id1, 2)));
    //     assert!(awset1.state.contains(&("apple".to_string(), node_id2, 3)));
    //     assert!(!awset2.state.contains(&("apple".to_string(), node_id2, 3)));
    //     assert!(awset2.context.contains(&(node_id2, 3)));
    // }

    // //Test ShoppingList
    // #[test]
    // fn test_add_new_item() {
    //     let mut list = ShoppingList::new();
    //     let node_id = Uuid::new_v4();
    //     list.add_or_update_item("Apple".to_string(), node_id, 5);
    //     assert_eq!(list.get_items().contains(&"Apple".to_string()), true);
    //     assert_eq!(list.items.get("Apple").unwrap().get_quantity(), 5);
    // }

    // #[test]
    // fn test_update_existing_item() {
    //     let mut list = ShoppingList::new();
    //     let node_id = Uuid::new_v4();
    //     list.add_or_update_item("Apple".to_string(), node_id, 5);
    //     list.add_or_update_item("Apple".to_string(), node_id, 3);
    //     assert_eq!(list.get_items().contains(&"Apple".to_string()), true);
    //     assert_eq!(list.items.get("Apple").unwrap().get_quantity(), 8);
    // }

    // #[test]
    // fn test_add_negative_quantity() {
    //     let mut list = ShoppingList::new();
    //     let node_id = Uuid::new_v4();
    //     list.add_or_update_item("Apple".to_string(), node_id, -3);
    //     assert_eq!(list.get_items().contains(&"Apple".to_string()), true);
    //     assert_eq!(list.items.get("Apple").unwrap().get_quantity(), -3);
    // }

    // #[test]
    // fn test_remove_item() {
    //     let mut list = ShoppingList::new();
    //     let node_id = Uuid::new_v4();
    //     list.add_or_update_item("Apple".to_string(), node_id, 5);
    //     list.remove_item("Apple".to_string());
    //     assert_eq!(list.get_items().contains(&"Apple".to_string()), false);
    // }

    // #[test]
    // fn test_merge_shopping_lists() {
    //     let mut list1 = ShoppingList::new();
    //     let mut list2 = ShoppingList::new();
    //     let node_id = Uuid::new_v4();

    //     list1.add_or_update_item("Apple".to_string(), node_id, 5);
    //     list2.add_or_update_item("Apple".to_string(), node_id, 3);

    //     list1.merge(list2);
    //     assert_eq!(list1.get_items().contains(&"Apple".to_string()), true);
    //     assert_eq!(list1.items.get("Apple").unwrap().get_quantity(), 8);
    // }


    // #[test]
    // fn test_get_items() {
    //     let mut list = ShoppingList::new();
    //     let node_id = Uuid::new_v4();
    //     list.add_or_update_item("Apple".to_string(), node_id, 5);
    //     list.add_or_update_item("Banana".to_string(), node_id, 2);

    //     let items = list.get_items();
    //     assert_eq!(items.len(), 2);
    //     assert!(items.contains(&"Apple".to_string()));
    //     assert!(items.contains(&"Banana".to_string()));
    // }



}


// #[cfg(test)]
// mod property_bounded_pn_counter_v2 {
//     use crate::crdt::crdt::*;
//     use uuid::Uuid;
//     use proptest::prelude::*;
//     use proptest::collection::vec;
//     use rand::random;
//     //Here we implement a strategy to generate random Uuids, because crate uuid doesnt have the Arbitrary trait required by proptest to generate random instances of Uuid for testing
//     fn uuid_strategy() -> impl Strategy<Value = Uuid> {
//         any::<[u8; 16]>().prop_map(Uuid::from_bytes)
//     }

//     fn bounded_pn_counter_strategy_v2() -> impl Strategy<Value = BoundedPNCounterv2> {
//         proptest::collection::vec((any::<bool>(), uuid_strategy(), any::<u32>()), 0..=10000)
//             .prop_map(|operations| {
//                 let mut counter = BoundedPNCounterv2::new();
                  

                
//                 for (increment, uuid, amount) in &operations {
//                     if *increment {
//                         counter.increment(*uuid, *amount);
//                     } else {
//                         counter.decrement(*uuid, *amount);
//                     }
//                 }
                
//                 counter
//             })
//     }

    
    
//     fn generate_operations(node_ids: &[Uuid]) -> Vec<(Uuid, bool, u32)> {
//         let mut operations = Vec::new();
//         for &node_id in node_ids {
//             let inc_or_dec = random::<bool>();
//             let amount = random::<u32>(); // Random amount between 0 and 99
//             operations.push((node_id, inc_or_dec, amount));
//         }
//         operations
//     }
    
//     fn apply_operations(ops: Vec<(Uuid, bool, u32)>) -> BoundedPNCounterv2 {
//         let mut counter = BoundedPNCounterv2::new();
//         for (node_id, increment, amount) in ops {
//             if increment {
//                 counter.increment(node_id, amount);
//             } else {
//                 counter.decrement(node_id, amount);
//             }
//         }
//         counter
//     }
    
    
    
//     fn non_comparable_crtd_strategy() -> impl Strategy<Value = (BoundedPNCounterv2, BoundedPNCounterv2)> {
//         let node_ids = proptest::collection::vec(uuid_strategy(), 1..10000); 

//         node_ids.prop_flat_map(|ids| {
//             let operations_a = generate_operations(&ids);
//             let operations_b = generate_operations(&ids);

//             (Just(apply_operations(operations_a)), Just(apply_operations(operations_b)))
//         })
//     }
    
//     proptest! {

//         #![proptest_config(ProptestConfig::with_cases(1000))]


//         #[test]
//         fn test_associativity(a in bounded_pn_counter_strategy_v2(), 
//                               b in bounded_pn_counter_strategy_v2(), 
//                               c in bounded_pn_counter_strategy_v2()) {

            
//             let ab_c = a.clone().merge(&b.clone()).merge(&c.clone());
//             let a_bc = a.merge(&b.merge(&c));
//             prop_assert_eq!(ab_c.positive_count, a_bc.positive_count);
//             prop_assert_eq!(ab_c.negative_count, a_bc.negative_count);
//         }
    
//         #[test]
//         fn test_commutativity(a in bounded_pn_counter_strategy_v2(), 
//                               b in bounded_pn_counter_strategy_v2()) {
//             let ab = a.clone().merge(&b.clone());
//             let ba = b.merge(&a);
//             prop_assert_eq!(ab.positive_count, ba.positive_count);
//             prop_assert_eq!(ab.negative_count, ba.negative_count);
//         }
    
//         #[test]
//         fn test_idempotency(a in bounded_pn_counter_strategy_v2()) {
//             let aa = a.clone().merge(&a.clone());
//             prop_assert_eq!(a.positive_count, aa.positive_count);
//             prop_assert_eq!(a.negative_count, aa.negative_count);
//         }
    
//         #[test]
//         fn test_handling_decrement(node_id in uuid_strategy(), amount in any::<u32>()) {
//             let mut counter = BoundedPNCounterv2::new();
//             counter.decrement(node_id, amount);
//             prop_assert!(counter.get_count() >= 0);
//         }
//         //Cases where CRDTs need to be always comparable
//         #[test]
//         fn test_compare(a in bounded_pn_counter_strategy_v2()) {
//             let mut b = a.clone();
//             prop_assert!(a.compare(&b)); // exactly equal states
//             // Add random amounts to both positive and negative counts of `b`
//             for node_id in a.positive_count().keys().chain(a.negative_count().keys()) {
//                 let additional_amount1: u32 = random::<u32>(); 
//                 let additional_amount2: u32 = random::<u32>();
//                 b.increment(*node_id, additional_amount1);
//                 b.decrement(*node_id, additional_amount2);
                
                
//             }
//             // Invariant: `a` in this context will be always less or equal then `b` -> so it's comparable
//             prop_assert!(a.compare(&b));
//         }

//         //TODO: check this examples later with professor
//         // // Cases where a and b are not comparable: a have some a[i] > b[i] and b have some b[i] > a[i] ( for a and b positive and negative counters)
//         // #[test]
//         // fn test_non_comparable_crtds(mut crdt_a in bounded_pn_counter_strategy_v2()) {
//         //     let mut crdt_b = crdt_a.clone();
//         //     let node_ida = Uuid::new_v4();
//         //     let node_idb = Uuid::new_v4();
//         //     let additional_amount1: u32 = random::<u32>(); 
//         //     let additional_amount2: u32 = random::<u32>();
//         //     crdt_b.increment(node_idb, additional_amount1);
//         //     crdt_a.increment(node_ida, additional_amount2);
//         //     prop_assert!(!crdt_a.compare(&crdt_b) && !crdt_b.compare(&crdt_a));
//         // }


//         #[test]
//         fn test_overflow(node_id in uuid_strategy(), amount in any::<u32>()) {
//             let mut counter = BoundedPNCounterv2::new();
//             counter.increment(node_id, u32::MAX);
//             counter.increment(node_id, amount);
//             // Check if value is either u32::MAX or wrapped around
//             assert!(counter.positive_count().get(&node_id) == Some(&u32::MAX) || 
//                     counter.positive_count().get(&node_id).unwrap() < &u32::MAX);
//         }
//     }
// }

//TODO: property tests
#[cfg(test)]
mod property_optimized_awset {
    use crate::crdt::crdt::*;
    use uuid::Uuid;
    use proptest::prelude::*;
    use std::collections::HashSet;

    fn awset_strategy() -> impl Strategy<Value = AWSet> {
        let item_strategy = ".*".prop_map(|s| s.to_string()); // Generate random strings for items
        let node_id_strategy = any::<[u8; 16]>().prop_map(Uuid::from_bytes); // Generate random Uuids for node ids
        let counter_strategy = any::<u32>(); // Generate random u32 for counters
    
        let state_strategy = proptest::collection::vec((item_strategy.clone(), node_id_strategy.clone(), counter_strategy), 0..1000);
        let context_strategy = proptest::collection::vec((node_id_strategy, counter_strategy), 0..1000);
    
        (state_strategy, context_strategy).prop_map(|(state, context)| {
            let mut awset = AWSet::new();
            awset.state = state.into_iter().collect::<HashSet<_>>();
            awset.context = context.into_iter().collect::<HashSet<_>>();
            awset
        })
    }

    proptest! {

        #![proptest_config(ProptestConfig::with_cases(1000))]



        #[test]
        fn test_add_remove(item in ".*", node_id in any::<[u8; 16]>().prop_map(Uuid::from_bytes)) {
            let mut awset = AWSet::new();
            awset.add_i(item.clone(), node_id);

            prop_assert!(awset.contains(&item));
            awset.rmv_i(item.clone());
            prop_assert!(!awset.contains(&item));
        }

        #[test]
        fn test_associativity_after_add(a in awset_strategy(), b in awset_strategy(), c in awset_strategy(), item in ".*", node_id in any::<[u8; 16]>().prop_map(Uuid::from_bytes)) {
            let mut a = a;
            let mut b = b;
            let mut c = c;

            a.add_i(item.clone(), node_id);
            b.add_i(item.clone(), node_id);
            c.add_i(item.clone(), node_id);

            let mut ab_c = a.clone();
            let mut a_bc = a.clone();
            let mut b_clone = b.clone();

            ab_c.merge(&b);
            ab_c.merge(&c);

            b_clone.merge(&c);
            a_bc.merge(&b_clone);
            prop_assert_eq!(ab_c.state, a_bc.state);
            prop_assert_eq!(ab_c.context, a_bc.context);
            
        }

        #[test]
        fn test_commutativity_after_add(a in awset_strategy(), b in awset_strategy(), item in ".*", node_id in any::<[u8; 16]>().prop_map(Uuid::from_bytes)) {
            let mut a = a;
            let mut b = b;

            a.add_i(item.clone(), node_id);
            b.add_i(item.clone(), node_id);

            let mut ab = a.clone();
            let mut ba = b.clone();

            ab.merge(&b);
            ba.merge(&a);

            prop_assert_eq!(ab.state, ba.state);
            prop_assert_eq!(ab.context, ba.context);
        }

        #[test]
        fn test_idempotence_after_add(a in awset_strategy(), item in ".*", node_id in any::<[u8; 16]>().prop_map(Uuid::from_bytes)) {
            let mut a = a;
            a.add_i(item.clone(), node_id);

            let mut aa = a.clone();
            aa.merge(&a);

            prop_assert_eq!(a.state, aa.state);
            prop_assert_eq!(a.context, aa.context);
        }

        //Test add remove conflits on properties

        #[test]
        fn test_associativity_after_add_remove(a in awset_strategy(), b in awset_strategy(), c in awset_strategy(), item in ".*", node_id in any::<[u8; 16]>().prop_map(Uuid::from_bytes)) {
            let mut a = a;
            let mut b = b;
            let mut c = c;

            a.add_i(item.clone(), node_id);
            b.add_i(item.clone(), node_id);
            c.add_i(item.clone(), node_id);

            // Apply remove operation only in one of the awsets
            a.rmv_i(item.clone());
            

            let mut ab_c = a.clone();
            let mut a_bc = a.clone();
            let mut b_clone = b.clone();

            ab_c.merge(&b);
            ab_c.merge(&c);

            b_clone.merge(&c);
            a_bc.merge(&b_clone);

            prop_assert_eq!(ab_c.state, a_bc.state);
            prop_assert_eq!(ab_c.context, a_bc.context);
        }

        #[test]
        fn test_commutativity_after_add_remove(a in awset_strategy(), b in awset_strategy(), item in ".*", node_id in any::<[u8; 16]>().prop_map(Uuid::from_bytes)) {
            let mut a = a;
            let mut b = b;

            a.add_i(item.clone(), node_id);
            b.add_i(item.clone(), node_id);

            // Apply remove operation
            b.rmv_i(item.clone());
            

            let mut ab = a.clone();
            let mut ba = b.clone();

            ab.merge(&b);
            ba.merge(&a);

            prop_assert_eq!(ab.state, ba.state);
            prop_assert_eq!(ab.context, ba.context);
        }

        

        #[test]
        fn test_idempotence_after_add_remove(a in awset_strategy(), item in ".*", node_id in any::<[u8; 16]>().prop_map(Uuid::from_bytes)) {
            let mut a = a;
            a.add_i(item.clone(), node_id);

            // Apply remove operation
            a.rmv_i(item.clone());

            let mut aa = a.clone();
            aa.merge(&a);

            prop_assert_eq!(a.state, aa.state);
            prop_assert_eq!(a.context, aa.context);
        }

        // just random inputs to do property tests

        #[test]
        fn test_associativity(a in awset_strategy(), b in awset_strategy(), c in awset_strategy()) {
            let mut ab_c = a.clone();
            let mut a_bc = a.clone();
            
            let mut b_clone = b.clone();
            ab_c.merge(&b);
            ab_c.merge(&c);
    
            b_clone.merge(&c);
            a_bc.merge(&b_clone);
    
            prop_assert_eq!(ab_c.state, a_bc.state);
            prop_assert_eq!(ab_c.context, a_bc.context);
        }
    
        #[test]
        fn test_commutativity(a in awset_strategy(), b in awset_strategy()) {
            let mut ab = a.clone();
            let mut ba = b.clone();
    
            ab.merge(&b);
            ba.merge(&a);
    
            prop_assert_eq!(ab.state, ba.state);
            prop_assert_eq!(ab.context, ba.context);
        }
    
        #[test]
        fn test_idempotence(a in awset_strategy()) {
            let mut aa = a.clone();
            aa.merge(&a);
    
            prop_assert_eq!(a.state, aa.state);
            prop_assert_eq!(a.context, aa.context);
        }
    
        #[test]
        fn test_convergence(a in awset_strategy(), b in awset_strategy(), c in awset_strategy()) {
            let mut ab = a.clone();
            let mut ac = a.clone();
            
            ab.merge(&b);
            ac.merge(&c);
    
            let mut bc = b.clone();
            bc.merge(&c);
    
            ab.merge(&bc);
            ac.merge(&bc);
    
            prop_assert_eq!(ab.state, ac.state);
            prop_assert_eq!(ab.context, ac.context);
        }
    
        #[test]
        fn test_element_addition_removal(item in ".*", node_id in any::<[u8; 16]>().prop_map(Uuid::from_bytes)) {
            let mut awset = AWSet::new();
            awset.add_i(item.clone(), node_id);
    
            prop_assert!(awset.contains(&item));
            awset.rmv_i(item.clone());
            prop_assert!(!awset.contains(&item));
        }
    }
    

}