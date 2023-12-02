
pub mod crdt_other_versions {
    use uuid::Uuid;
    use std::collections::HashSet;
    use std::collections::HashMap;
    

    

    #[derive(Clone,Debug)]
    pub struct GCounter {
            positive_count: i32,
            
        }

    impl GCounter{
       pub fn new() -> Self {
            GCounter {
                positive_count: 0,
            }
        }
        pub fn get_positive_count(&self) -> i32 {
            self.positive_count
        }

        
        pub fn increment(&mut self, ammount: i32){
            self.positive_count += ammount;
        }

        

        pub fn get_count(&self) -> i32 {
            self.positive_count 
        }

        pub fn compare(&self, inc_pn_counter: &GCounter) -> bool{
            self.get_positive_count() <= inc_pn_counter.get_positive_count() 
        }
        // merge function perserving: commutative, associative, and idempotent.
        pub fn merge(&mut self, inc_pn_counter: &GCounter) {
            self.positive_count = std::cmp::max(self.positive_count, inc_pn_counter.positive_count);
        }
    }


    #[derive(Clone,Debug)]
   pub struct SimplePNCounter {
        positive_count: i32,
        negative_count: i32,
    }

    impl SimplePNCounter{
       pub fn new() -> Self {
            SimplePNCounter {
                positive_count: 0,
                negative_count: 0,
            }
        }
        pub fn get_positive_count(&self) -> i32 {
            self.positive_count
        }

        pub fn get_negative_count(&self) -> i32 {
            self.negative_count
        }
        pub fn increment(&mut self, ammount: i32){
            self.positive_count += ammount;
        }

        pub fn decrement(&mut self, ammount: i32){
            self.negative_count += ammount;
        }

        pub fn get_count(&self) -> i32 {
            self.positive_count - self.negative_count
        }

        pub fn compare(&self, inc_pn_counter: &SimplePNCounter) -> bool{
            self.get_positive_count() <= inc_pn_counter.get_positive_count() && self.get_negative_count() <= inc_pn_counter.get_negative_count()
        }
        // merge function perserving: commutative, associative, and idempotent.
        pub fn merge(&mut self, inc_pn_counter: &SimplePNCounter) {
            self.positive_count = std::cmp::max(self.positive_count, inc_pn_counter.positive_count);
            self.negative_count = std::cmp::max(self.negative_count, inc_pn_counter.negative_count);
        }
    }

    //TODO: Improve Pn-Counter using vector for values of all nod_id's ?
    #[derive(Clone,Debug)]
   pub struct BoundedPNCounter {
        pub positive_count: HashMap<Uuid, u32> ,
        pub negative_count: HashMap<Uuid, u32>,
    }

    impl BoundedPNCounter{
       pub fn new() -> Self {
            BoundedPNCounter {
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
        pub fn increment(&mut self, node_id: Uuid) {
            let inc_count = self.positive_count.entry(node_id).or_insert(0);
            *inc_count += 1;
        }
        pub fn decrement(&mut self, node_id: Uuid) {
            let dec_count = self.negative_count.entry(node_id).or_insert(0);
            let positive_count = self.positive_count.entry(node_id).or_insert(0);
    
            // Here we ensure the total count never goes below zero -> lower limit
            if *dec_count + 1 <= *positive_count {
                *dec_count += 1;
            } else {
                // Decrement would turn possible to have negative count/quantities, so by default count = 0 inc and dec will have same value
                *dec_count = *positive_count;
            }
        }
    

        pub fn get_count(&self) -> u32 {
            let sum_pos_count: u32 = self.positive_count.values().sum();
            let sum_neg_count: u32 = self.negative_count.values().sum();

            sum_pos_count - sum_neg_count
        
        }

        pub fn compare(&self, inc_b_pn_counter: &BoundedPNCounter) -> bool{
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
        pub fn merge(&self, other: &BoundedPNCounter) -> BoundedPNCounter {
            let mut merged = BoundedPNCounter::new();
    
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


    #[derive(Clone,Debug)]
    pub struct Item {
        // id: Uuid,
        name: String, // comparar por nome -> ver restrições a garantir

        quantity_counter: BoundedPNCounter, // Será a quantidade tendo em conta os: increments and decrements

    }

    impl Item {
        pub fn new( name: String) -> Self{
            Item {
                // id, check this later
                name,
                quantity_counter: BoundedPNCounter::new(),
            

            }

        }

        // pub fn get_id(&self) -> uuid::Uuid {
        //     self.id
        // }
    
        pub fn get_name(&self) -> &str {
            &self.name
        }

        pub  fn increment_quantity(&mut self, node_id: Uuid, increment: i32){
                self.quantity_counter.increment(node_id, increment);
            }
        pub  fn decrement_quantity(&mut self,node_id: Uuid ,decrement: i32){
                self.quantity_counter.decrement(node_id, decrement);
            }

        pub  fn get_quantity(&self) -> i32{
            return self.quantity_counter.get_count();
        }
        //Merge current item quantity with other item
        pub  fn merge(&mut self, incoming_item: &Item) {
            if self.name == incoming_item.name{
                self.quantity_counter.merge(&incoming_item.quantity_counter);
            }
        }
    }

    // Versão não optimizada em relação ao espaço utilizado pelo crdt AWSET
    #[derive(Clone, Debug)]
    pub struct AWSetV0 {
        pub state: HashSet<(String, Uuid, i32)>, // Set of tuples (Item, NodeId, Counter)
        pub context: HashSet<(Uuid, i32)>, // Set of tuples (NodeId, Counter)
    }
    impl AWSetV0 {
        pub fn new() -> Self {
            //TODO: check if Hashset is the best struture to use, do we need to save every 3-tuple and 2-tuple on state and context, or just the more recent one
            AWSetV0{
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

        
        //TODO: GCounter ou pertence ao AWSet, ou não é necessário
        // addi(e,(s, c)) -> e= name of created item, s= state, c = context
        // rmvi(e,(s, c)) -> e= name of created item, s= state, c = context
        //maxi(c) = find the max on context set
        // nexti(c) = create the next (NodeId, Counter) taken into account the existing ones on context = (NodeId, Counter) set
        //
        pub fn max_i(&self, node_id: Uuid) -> i32 {
            self.context.iter()
                .filter(|(uuid, _)| *uuid == node_id)
                .map(|(_, counter)| *counter)
                .max()
                .unwrap_or(0)//if there is no (NodeId, Counter) tuple, max_i returns 0, so nex_i can generate tuples(node_id, 0+1( Counter)), when there is no pair on the context
        }
    
        pub fn next_i(&self, node_id: Uuid) -> (Uuid, i32) {
            (node_id, self.max_i(node_id) + 1) 
        }
        //TODO: we really need to ahve all the (item_name,Nodeid, Counter_i), or when we update causal context, we also just update (item_name,Nodeid, Counter_i)
        // addd new item or increment/decrement existing item: +x increment, -x decrement
        pub fn add_i(&mut self, item_name: String, node_id: Uuid) {
            let next_context = Self::next_i(&self, node_id);
            self.context.insert(next_context.clone()); // c ∪ {d}

            
            
            // just update item without quantity and or return updated_item change with updated next_i
            self.state.insert((item_name.clone(), next_context.0, next_context.1 )); // s ∪ {(e, d)}

        }
        //Here we remove all tuples from state, that have the item_name
        pub fn rmv_i(&mut self, item_name: String) {
            self.state.retain(|(name,_,_)| *name != item_name);
            
        }

        pub fn filter(&self, inc_awset: &AWSetV0) -> HashSet<(String, Uuid, i32)> {
            self.state.iter()
            .filter(|(_name, node_id, counter)| {
                !inc_awset.context.iter().any(|(inc_node_id,inc_counter)|{
                    node_id == inc_node_id && counter < inc_counter 
                })
            })
            .cloned()
            .collect()
        }

        pub fn merge(&mut self, inc_awset: &AWSetV0){
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

    // For this version, we improve using Maps and reducing the data needed on CRDT, saving only the most updated last (node_id,Counter1) in state and (node_id, Counter2) in context
    #[derive(Clone, Debug)]
    pub struct AWSetOptV2 {
        // state: Map(item_name, Map(NodeId, Counter))
        pub state: HashMap<String, HashMap<Uuid, u32>>,
        // map : Map(node_id, Counter)
        pub context: HashMap<Uuid, u32>,
    }

    impl AWSetOptV2 {
        pub fn new() -> Self {
            AWSetOptV2 {
                state: HashMap::new(),
                context: HashMap::new(),
            }
        }

        pub fn elements(&self) -> Vec<String> {
            self.state.keys().cloned().collect()
        }

        pub fn contains(&self, item_name: &str) -> bool {
            self.state.contains_key(item_name)
        }
        //Max will be always the value saved on context for a given node_id
        pub fn max_i(&self, node_id: Uuid) -> u32 {
            *self.context.get(&node_id).unwrap_or(&0)
        }

        pub fn next_i(&self, node_id: Uuid) -> (Uuid, u32) {
            (node_id, self.max_i(node_id) + 1)
        }

        
        pub fn add_i(&mut self, item_name: String, node_id: Uuid) {
            let next_context = Self::next_i(&self, node_id);
            self.context.insert(next_context.0, next_context.1);

            self.state
                .entry(item_name.clone())
                .or_insert_with(HashMap::new)
                .insert(next_context.0, next_context.1);
        }

        pub fn rmv_i(&mut self, item_name: String) {
            self.state.remove(&item_name);
        }

        pub fn filter(&self, inc_awset: &AWSetOptV2) -> HashMap<String, HashMap<Uuid, u32>> {
            self.state.iter()
                .filter(|(name, node_map)| {
                    node_map.iter().any(|(node_id, counter)| {
                        !inc_awset.context.get(node_id).map_or(false, |inc_counter| counter < inc_counter)
                    })
                })
                .map(|(name, node_map)| (name.clone(), node_map.clone()))
                .collect()
        }

        pub fn merge(&mut self, inc_awset: &AWSetOptV2) {
            // Intersection between states of two AWSetOptV2
            let states_intersection = self.state.keys()
                .filter(|key| inc_awset.state.contains_key(*key))
                .map(|key| key.clone())
                .collect::<HashSet<String>>();

            // Union of filter(s,c') U f(s',c)
            let filter_state_1 = self.filter(inc_awset);
            let filter_state_2 = inc_awset.filter(self);

            // Get intersection of states s and s'
            let mut final_merge = HashMap::new();
            for key in states_intersection {
                if let (Some(self_map), Some(inc_map)) = (self.state.get(&key), inc_awset.state.get(&key)) {
                    let mut merged_map = self_map.clone();
                    for (node_id, counter) in inc_map {
                        merged_map.insert(*node_id, *counter);
                    }
                    final_merge.insert(key, merged_map);
                }
            }

            for (key, node_map) in filter_state_1.into_iter().chain(filter_state_2) {
                final_merge.insert(key, node_map);
            }

            // Union of contexts
            let final_context = self.context.iter()
                .chain(inc_awset.context.iter())
                .map(|(&k, &v)| (k, v))
                .collect();

            // Update state and context
            self.state = final_merge;
            self.context = final_context;
        }
    }

}

pub mod tests_other_versions {
    use crate::crdt_other_versions::crdt_other_versions::*;
    use uuid::Uuid;
    use std::collections::HashSet;

    #[test]
    fn test_new() {
        let counter = BoundedPNCounter::new();
        assert!(counter.positive_count.is_empty());
        assert!(counter.negative_count.is_empty());
    }

    #[test]
    fn test_increment() {
        let mut counter = BoundedPNCounter::new();
        let node_id = Uuid::new_v4();
        counter.increment(node_id);
        assert_eq!(*counter.positive_count.get(&node_id).unwrap(), 1);
    }

    #[test]
    fn test_decrement() {
        let mut counter = BoundedPNCounter::new();
        let node_id = Uuid::new_v4();
        counter.increment(node_id);
        counter.decrement(node_id);
        assert_eq!(*counter.negative_count.get(&node_id).unwrap(), 1);
        counter.decrement(node_id); // Should not decrease further
        assert_eq!(*counter.negative_count.get(&node_id).unwrap(), 1);
    }

    #[test]
    fn test_get_count() {
        let mut counter = BoundedPNCounter::new();
        let node_id = Uuid::new_v4();
        counter.increment(node_id);
        counter.increment(node_id);
        counter.decrement(node_id);
        assert_eq!(counter.get_count(), 1);
    }

    #[test]
    fn test_compare() {
        let mut counter1 = BoundedPNCounter::new();
        let mut counter2 = BoundedPNCounter::new();
        let node_id = Uuid::new_v4();
        counter1.increment(node_id);
        counter2.increment(node_id);
        assert!(counter1.compare(&counter2));
        counter2.increment(node_id);
        assert!(counter1.compare(&counter2));
        //Test an increment on counter1 with different node_id that counter2 dont have 
        let node_id2 = Uuid::new_v4();
        counter1.increment(node_id2);
        assert!(!counter1.compare(&counter2));
    }

    #[test]
    fn test_merge_commutativity() {
        let mut counter1 = BoundedPNCounter::new();
        let mut counter2 = BoundedPNCounter::new();
        let node_id = Uuid::new_v4();
        counter1.increment(node_id);
        counter2.decrement(node_id);
        
        let merged1 = counter1.merge(&counter2);
        let merged2 = counter2.merge(&counter1);
        assert_eq!(merged1.positive_count, merged2.positive_count);
        assert_eq!(merged1.negative_count, merged2.negative_count);
    }

    #[test]
    fn test_merge_associativity() {
        let mut counter1 = BoundedPNCounter::new();
        let mut counter2 = BoundedPNCounter::new();
        let mut counter3 = BoundedPNCounter::new();
        let node_id = Uuid::new_v4();
        counter1.increment(node_id);
        counter2.decrement(node_id);
        counter3.increment(node_id);

        let merged1 = counter1.merge(&counter2).merge(&counter3);
        let merged2 = counter1.merge(&counter2.merge(&counter3));
        assert_eq!(merged1.positive_count, merged2.positive_count);
        assert_eq!(merged1.negative_count, merged2.negative_count);
    }

    #[test]
    fn test_merge_idempotency() {
        let mut counter = BoundedPNCounter::new();
        let node_id = Uuid::new_v4();
        counter.increment(node_id);

        let merged = counter.merge(&counter);
        
        assert_eq!(counter.positive_count, merged.positive_count);
        assert_eq!(counter.negative_count, merged.negative_count);
    }


    // //Test AWSetV0

   

    #[test]
    fn test_awset_new() {
        let awset = AWSetV0::new();
        assert!(awset.state.is_empty());
        assert!(awset.context.is_empty());
    }

    #[test]
    fn test_max_i() {
        let mut awset = AWSetV0::new();
        let node_id = Uuid::new_v4();
        awset.context.insert((node_id, 1));
        awset.context.insert((node_id, 3));
        awset.context.insert((node_id, 2));

        assert_eq!(awset.max_i(node_id), 3);
    }

    #[test]
    fn test_next_i() {
        let mut awset = AWSetV0::new();
        let node_id = Uuid::new_v4();
        awset.context.insert((node_id, 1));
        awset.context.insert((node_id, 2));

        let next = awset.next_i(node_id);
        assert_eq!(next, (node_id, 3));
    }


    #[test]
    fn test_context_with_multiple_nodes() {
        let mut awset = AWSetV0::new();
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
        let mut awset = AWSetV0::new();
        let node_id = Uuid::new_v4();
        let item_name = "apple".to_string();
        

        awset.add_i(item_name.clone(), node_id);
        
       
        assert!(awset.state.contains(&(item_name, node_id, 1)));
        assert!(awset.context.contains(&(node_id, 1)));
    }

    #[test]
    fn test_increment_existing_item() {
        let mut awset = AWSetV0::new();
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
        let mut awset = AWSetV0::new();
        let node_id = Uuid::new_v4();
        let item_name = "apple".to_string();
        awset.state.insert((item_name.clone(), node_id, 1));
        awset.context.insert((node_id, 1));

        awset.add_i(item_name.clone(), node_id);
        
        
        assert!(awset.state.contains(&(item_name, node_id, 2)));
        assert!(awset.context.contains(&(node_id, 2)));
    }

    fn test_add_i() {
        let mut awset = AWSetV0::new();
        let node_id = Uuid::new_v4();
        let item_name = "apple".to_string();

       
        awset.add_i(item_name.clone(), node_id);
        assert_eq!(awset.state.contains(&(item_name.clone(), node_id, 1)), true);
    }

    // Unit tests for rmv_i
    #[test]
    fn test_rmv_i_existing_item() {
        let mut awset = AWSetV0::new();
        let node_id = Uuid::new_v4();
        let item_name = "apple".to_string();

        
        awset.add_i(item_name.clone(), node_id);
        awset.rmv_i(item_name.clone());

        assert_eq!(awset.state.contains(&(item_name, node_id, 1)), false);
    }

    
    #[test]
    fn test_rmv_i_non_existent_item() {
        let mut awset = AWSetV0::new();
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
        let mut awset = AWSetV0::new();
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
        let mut awset_1 = AWSetV0::new();
        let mut awset_2 = AWSetV0::new();

        
        awset_1.state.insert(("apple".to_string(), node_id_1, 1));
        awset_1.context.insert((node_id_1, 2));

        awset_2.state.insert(("banana".to_string(), node_id_2, 1));
        awset_2.context.insert((node_id_2, 2));

        // Expected result after filtering awset_1 against awset_2: Mock
        let mut expected_state: HashSet<(String, Uuid, i32)> = HashSet::new(); //"apple" should not be in the filtered state
        expected_state.insert(("apple".to_string(), node_id_1, 1));
        
        let filtered_state = awset_1.filter(&awset_2);

        // Check that the filtered state matches the expected state
        assert_eq!(filtered_state, expected_state);
    }

    //Test merge

    #[test]
    fn test_merge_with_overlap() {
        let mut awset1 = AWSetV0::new();
        let mut awset2 = AWSetV0::new();

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
        let mut awset1 = AWSetV0::new();
        let mut awset2 = AWSetV0::new();

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
        let mut awset1 = AWSetV0::new();
        let awset2 = AWSetV0::new();

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
        let mut awset1 = AWSetV0::new();
        let awset2 = AWSetV0::new();

        // Merging two empty sets should result in an empty set
        awset1.merge(&awset2);
        assert!(awset1.state.is_empty());
    }

    #[test]
    fn test_elements() {
        let mut awset = AWSetV0::new();
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
        let mut awset = AWSetV0::new();
        let node_id = Uuid::new_v4();

        // Adding items
        awset.add_i("Apple".to_string(), node_id);

        assert!(awset.contains("Apple"));
        assert!(!awset.contains("Banana"));
    }

    //Test AWSetOptV2

    #[test]
    fn test_new_awsetoptv2() {
        let awset = AWSetOptV2::new();
        assert!(awset.state.is_empty());
        assert!(awset.context.is_empty());
    }

    #[test]
    fn test_add_and_contains_elements() {
        let mut awset = AWSetOptV2::new();
        let node_id = Uuid::new_v4();
        awset.add_i("Apple".to_string(), node_id);

        assert!(awset.contains("Apple"));
        assert!(!awset.contains("Banana"));
    }

    #[test]
    fn test_remove_element() {
        let mut awset = AWSetOptV2::new();
        let node_id = Uuid::new_v4();
        awset.add_i("Apple".to_string(), node_id);
        awset.rmv_i("Apple".to_string());

        assert!(!awset.contains("Apple"));
    }

    #[test]
    fn test_elements_list() {
        let mut awset = AWSetOptV2::new();
        let node_id = Uuid::new_v4();
        awset.add_i("Apple".to_string(), node_id);
        awset.add_i("Banana".to_string(), node_id);

        let elements = awset.elements();
        assert_eq!(elements.len(), 2);
        assert!(elements.contains(&"Apple".to_string()));
        assert!(elements.contains(&"Banana".to_string()));
    }

    #[test]
    fn test_merge_awsetoptv2() {
        let mut awset1 = AWSetOptV2::new();
        let mut awset2 = AWSetOptV2::new();
        let node_id1 = Uuid::new_v4();
        let node_id2 = Uuid::new_v4();

        awset1.add_i("Apple".to_string(), node_id1);
        awset2.add_i("Banana".to_string(), node_id2);

        awset1.merge(&awset2);

        assert!(awset1.contains("Apple"));
        assert!(awset1.contains("Banana"));
        assert_eq!(awset1.state.len(), 2);
    }

    // More advanced tests for AWSetOptV2

    #[test]
    fn test_merge_with_conflicts() {
        let mut awset1 = AWSetOptV2::new();
        let mut awset2 = AWSetOptV2::new();
        let node_id1 = Uuid::new_v4();
        let node_id2 = Uuid::new_v4();

        
        awset1.add_i("Apple".to_string(), node_id1);
        awset2.add_i("Apple".to_string(), node_id2);
        awset2.add_i("Apple".to_string(), node_id2);
        let counter0 = awset1.state.get("Apple").unwrap().get(&node_id1).unwrap_or(&0);
        assert_eq!(*counter0, 1);

        awset1.merge(&awset2);

        assert!(awset1.contains("Apple"));
        assert_eq!(awset1.state.get("Apple").unwrap().len(), 1);

        
        let counter1 = awset1.state.get("Apple").unwrap().get(&node_id2).unwrap_or(&0);
        let counter2 = awset2.state.get("Apple").unwrap().get(&node_id2).unwrap_or(&0);
        let counter3 = awset1.state.get("Apple").unwrap().get(&node_id1).unwrap_or(&0);
        assert_eq!(counter1, counter2);
        // assert_eq!(*counter3, 1);
    }

    #[test]
    fn test_merge_with_removal() {
        let mut awset1 = AWSetOptV2::new();
        let mut awset2 = AWSetOptV2::new();
        let node_id = Uuid::new_v4();

        awset1.add_i("Apple".to_string(), node_id);
        awset2.add_i("Apple".to_string(), node_id);
        awset2.rmv_i("Apple".to_string());

        awset1.merge(&awset2);

        assert!(awset1.contains("Apple"));
    }


    #[test]
    fn test_filter_without_overlapping_node_ids() {
        let mut awset1 = AWSetOptV2::new();
        let mut awset2 = AWSetOptV2::new();
        let node_id1 = Uuid::new_v4();
        let node_id2 = Uuid::new_v4();

        // Populate awset1
        awset1.add_i("Apple".to_string(), node_id1);
        awset1.add_i("Banana".to_string(), node_id1);

        // Populate awset2 with different node_id
        awset2.add_i("Apple".to_string(), node_id2);
        awset2.add_i("Banana".to_string(), node_id2);

        let filtered = awset1.filter(&awset2);
        assert_eq!(filtered.len(), 2); // No entries should be filtered out
    }

    #[test]
    fn test_filter_with_overlapping_node_ids_and_different_counters() {
        let mut awset1 = AWSetOptV2::new();
        let mut awset2 = AWSetOptV2::new();
        let node_id = Uuid::new_v4();

        // Populate both sets with the same node_id but different counters
        awset1.add_i("Apple".to_string(), node_id);
        awset1.add_i("Banana".to_string(), node_id);
        awset2.add_i("Apple".to_string(), node_id); // This will increment the counter in awset2

        let filtered = awset1.filter(&awset2); // Banana is not known in awset2, should remain in filtered
        assert_eq!(filtered.len(), 1); 
        assert!(filtered.contains_key("Banana"));
    }




    // Test Commutativity: merge(a, b) == merge(b, a)
    #[test]
    fn test_commutativity_of_merge() {
        let mut awset1 = AWSetOptV2::new();
        let mut awset2 = AWSetOptV2::new();
        let node_id = Uuid::new_v4();

        awset1.add_i("Apple".to_string(), node_id);
        awset2.add_i("Banana".to_string(), node_id);

        let mut awset1_clone = awset1.clone();
        let mut awset2_clone = awset2.clone();

        awset1.merge(&awset2);
        awset2_clone.merge(&awset1_clone);

        assert_eq!(awset1.state, awset2_clone.state);
    }

    // Test Associativity: merge(a, merge(b, c)) == merge(merge(a, b), c)
    #[test]
    fn test_associativity_of_merge() {
        let mut awset1 = AWSetOptV2::new();
        let mut awset2 = AWSetOptV2::new();
        let mut awset3 = AWSetOptV2::new();
        let node_id = Uuid::new_v4();

        awset1.add_i("Apple".to_string(), node_id);
        awset2.add_i("Banana".to_string(), node_id);
        awset3.add_i("Cherry".to_string(), node_id);

        let mut left_merge = awset1.clone();
        let mut right_merge = awset2.clone();
        let mut awset2_clone = awset2.clone();

        left_merge.merge(&awset2_clone);
        left_merge.merge(&awset3);

        right_merge.merge(&awset3);
        awset1.merge(&right_merge);

        assert_eq!(left_merge.state, awset1.state);
    }

    // Test Idempotency: merge(a, a) == a
    #[test]
    fn test_idempotency_of_merge() {
        let mut awset1 = AWSetOptV2::new();
        let node_id = Uuid::new_v4();

        awset1.add_i("Apple".to_string(), node_id);

        let awset1_clone = awset1.clone();
        awset1.merge(&awset1_clone);

        assert_eq!(awset1.state, awset1_clone.state);
    }

}


#[cfg(test)]
mod property_bounded_pn_counter {
    use crate::crdt::crdt::*;
    use uuid::Uuid;
    use proptest::prelude::*;
    use proptest::collection::vec;
    //Here we implement a strategy to generate random Uuids, because crate uuid doesnt have the Arbitrary trait required by proptest to generate random instances of Uuid for testing
    fn uuid_strategy() -> impl Strategy<Value = Uuid> {
        any::<[u8; 16]>().prop_map(Uuid::from_bytes)
    }

    fn counter_strategy() -> impl Strategy<Value = Vec<(bool, Uuid)>> {
        proptest::collection::vec((any::<bool>(), uuid_strategy()), 0..=10000)
    }

    proptest! {



        #![proptest_config(ProptestConfig::with_cases(1000))]

        #[test]
        fn prop_test_increment_decrement_consistency(node_id in uuid_strategy(), inc_times in 0u32..10, dec_times in 0u32..10) {
            let mut counter = BoundedPNCounter::new();

            for _ in 0..inc_times {
                counter.increment(node_id);
            }

            
            for _ in 0..dec_times {
                counter.decrement(node_id);
            }
            let sum_pos_count: u32 = counter.positive_count.values().sum();
            let sum_neg_count: u32 = counter.negative_count.values().sum();
            prop_assert_eq!(counter.get_count(), sum_pos_count - sum_neg_count);
        }

        #[test]
        fn test_decrement_handling(
            node_id in uuid_strategy(),
            increment_count in 0u32..=1000,
            decrement_count in 1001u32..=2000 // We will have more decrements than increments
        ) {
            let mut crdt = BoundedPNCounter::new();

           
            for _ in 0..increment_count {
                crdt.increment(node_id);
            }

            
            for _ in 0..decrement_count {
                crdt.decrement(node_id);
            }

            // Check if the decrement function handled the situation correctly
            let final_positive_count = *crdt.positive_count().get(&node_id).unwrap_or(&0);
            let final_negative_count = *crdt.negative_count().get(&node_id).unwrap_or(&0);

            //Invariant:  negative count never exceeds the positive count
            prop_assert!(final_negative_count <= final_positive_count, 
                "Negative count should not exceed positive count: Negative = {}, Positive = {}", final_negative_count, final_positive_count);

            // Invariant: total count  is always non-negative
            let total_count = final_positive_count as i32 - final_negative_count as i32;
            prop_assert!(total_count >= 0, "Total count should never be negative: {}", total_count);
        }

        #[test]
        fn prop_test_compare_consistency(node_id in uuid_strategy(), inc_times in 0u32..10) {
            let mut counter = BoundedPNCounter::new();

            for _ in 0..inc_times {
                counter.increment(node_id);
            }

            prop_assert!(counter.compare(&counter));
        }

        fn prop_test_compare_property(
            operations in counter_strategy(),
            extra_operations in counter_strategy(),
        ) {
            let mut crdt1 = BoundedPNCounter::new();
            let mut crdt2 = BoundedPNCounter::new();
    
            
            for (increment, node_id) in &operations {
                if *increment {
                    crdt1.increment(*node_id);
                    crdt2.increment(*node_id);
                } else {
                    crdt1.decrement(*node_id);
                    crdt2.decrement(*node_id);
                }
            }
            prop_assert!(crdt1.compare(&crdt2)); // invariant: all Equal between states of crdt1 and crdt2
            // Applying extra operations to the second CRDT
            for (increment, node_id) in &extra_operations {
                if *increment {
                    crdt2.increment(*node_id);
                } else {
                    crdt2.decrement(*node_id);
                }
            }
    
            // Invariant: Assert that crdt1 compares as less than or equal to crdt2: 
            // In the conditions above of our test: crdt1 needs to be  always subset of crdt2 -> compare needs always to return true
            prop_assert!(crdt1.compare(&crdt2));
        }

        // Here we test the properties of Idempotency, commutative, associative, also the limit of minimum of value == 0 ( inc_value -dec_value) is tested

        #[test]
        fn prop_test_merge_idempotent(ops in counter_strategy(),inc_times in 0u32..10) {
            let mut merged1 = BoundedPNCounter::new();
            
            for _ in 0..inc_times {
            
                // Apply operations to counter
                for (increment, node_id) in &ops {
                    if *increment {
                        merged1.increment(*node_id);
                    } else {
                        merged1.decrement(*node_id);
                    }
                }
            }
            // Idempotence: crdt merged with itself should equal itself
            let merged_self = merged1.merge(&merged1);
            //Invariant: same states : ( the inc and dec counters for all node_ids)
            prop_assert_eq!(merged1.positive_count, merged_self.positive_count);
            prop_assert_eq!(merged1.negative_count, merged_self.negative_count);
        }
        #[test]
        fn prop_test_merge_commutative(ops1 in counter_strategy(), ops2 in counter_strategy(),inc_times in 0u32..10) {
            let mut counter1 = BoundedPNCounter::new();
            let mut counter2 = BoundedPNCounter::new();
            
            for _ in 0..inc_times {
                // Apply different operations to both counters
                for (increment, node_id) in &ops1 {
                    if *increment {
                        counter1.increment(*node_id);
                        
                    } else {
                        counter1.decrement(*node_id);
                    
                    }
                }
                for (increment, node_id) in &ops2 {
                    if *increment {
                        
                        counter2.increment(*node_id);
                    } else {
                        
                        counter2.decrement(*node_id);
                    }
                }

            }
    
            // Commutativity: crdt merged with counter2 should equal counter2 merged with counter1
            let merged1 = counter1.merge(&counter2);
            let merged2 = counter2.merge(&counter1);
            //Invariant: same states
            prop_assert_eq!(merged1.positive_count, merged2.positive_count);
            prop_assert_eq!(merged1.negative_count, merged2.negative_count);
        }

        #[test]
        fn prop_test_merge_associative(ops1 in counter_strategy(), ops2 in counter_strategy(), ops3 in counter_strategy(),inc_times in 0u32..10 ) {
            let mut counter1 = BoundedPNCounter::new();
            let mut counter2 = BoundedPNCounter::new();
            let mut counter3 = BoundedPNCounter::new();

            // Apply different operations to 3 CRDTs
            for _ in 0..inc_times {
                for (increment, node_id) in &ops1 {
                    if *increment {
                        counter1.increment(*node_id);
                        
                    } else {
                        counter1.decrement(*node_id);
                        
                    }
                }
                for (increment, node_id) in &ops2 {
                    if *increment {
                        
                        counter2.increment(*node_id);
                        
                    } else {
                        
                        counter2.decrement(*node_id);
                        
                    }
                }
                for (increment, node_id) in &ops3 {
                    if *increment {
                        
                        counter3.increment(*node_id);
                    } else {
                        
                        counter3.decrement(*node_id);
                    }
                }
            
            }
            // Associativity: (crdt merged with counter2) merged with counter3 should equal crdt merged with (counter2 merged with counter3)
            let merged1_then_3 = counter1.merge(&counter2).merge(&counter3);
            let merged2_then_3 = counter1.merge(&counter2.merge(&counter3));
            //invariant: equal states
            prop_assert_eq!( merged1_then_3.positive_count, merged2_then_3.positive_count);
            prop_assert_eq!( merged1_then_3.negative_count, merged2_then_3.negative_count);
        }

        
    }
}