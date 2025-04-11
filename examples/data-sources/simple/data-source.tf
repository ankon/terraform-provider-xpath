data "xpath_query_one" "load_from_feed" {
  content    = file("feed.xml")
  expression = "//"
  namespace_bindings = {
    atom    = "http://www.w3.org/2005/Atom"
    pingdom = "http://www.pingdom.com/ns/PingdomRSSNamespace"
  }
}