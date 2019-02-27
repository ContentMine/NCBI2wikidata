Tool for finding open access PubMed data and generating wikidata updates from that
---------------

This tool takes in a list of subject terms to search PubMed for, and outputs a set of updates to be applied to wikidata from that source.

The tool generates two files on each run: `results.csv` is a human readable version of the data collected from PubMed, and `results_quickstatements.txt` is a valid quickstatements file containing the updates such that they can be applied to wikidata.

To run the tool you need two things:

* You need a valid NCBI API KEY. You can generate one of these by logging into NCBI and finding the API key section on your settings page.
* You'll need a JSON file that is an array of subjects to search for. E.g.,:

```
[
    "Leptospirosis",
    "Rett Syndrome"
]
```

Information sources
==================

This tool primarily gets the information by looking up reviewed publications on PubMed using the search term:

"Subject[Mesh Major Topic] AND Review[ptyp]"

That is to say, it looks up each subject in term to see which papers list it as a Major Topic, and we restrict our publication type to reviews.

In addition to PubMed, the tool will lookup more detailed license information from EuroPMC where available, as the PubMed open access license information lacks detailed versions.



Building
===========

NCBI2wikidata is written in Go, and built with Make. You also need to set GOPATH to the directory of the project. If you're in the root directory of this source tree then you can simply type:

```export GOPATH=$PWD```

You will need to fetch the libraries that this depends on, which you can do with:

```make get```

Then you should be able to just run:

```make```

And the tool will be built and put into the `$GOPATH/bin` directory.


Generating the feed
===================

There is a tool that you can use to generate a list of potential search terms. It will write out a JSON file of the top diseases from a provided the list health specialities. The speciality list must be provided as an input JSON file of the form:

```
[
    {
        "spec":"http://www.wikidata.org/entity/Q27721906",
        "specLabel":"Oral and maxillofacial surgery"
    },
    {
        "spec":"http://www.wikidata.org/entity/Q27712466",
        "specLabel":"Oral surgery, oral medicine, and oral pathology"
    }
]
```

Which could be generated on wikidata with a SPARQL query like so:

```
SELECT ?spec ?specLabel WHERE
{
  {
    SELECT ?spec (COUNT(?item) AS ?count) WHERE {
        ?item wdt:P31 wd:Q12136 .
        ?item wdt:P1995 ?spec  .
        }
  GROUP BY ?spec
  }
   SERVICE wikibase:label { bd:serviceParam wikibase:language "en" }
} LIMIT 2
```

And then you save the result as non-verbose JSON.

The tool itself is then invoked like so:

```bin/GenerateMeshTerms -feed specialities.json```

It will take a while to run if you provide many specialities due to the limit at which NCBI APIs can be called. Once complete it will have created a file called `generated_feed.json`.  This file can then be passed to the NSCBI2wikidata tool.


License
============

This software is copyright Content Mine Ltd 2019, and released under the Apache 2.0 License.

This software calls the NCBI PubMed API, which is subject to the following disclaimer and copyright notice: https://www.ncbi.nlm.nih.gov/home/about/policies/

Dependencies
============

Relies on

* https://github.com/ContentMine/wikibase
* https://github.com/ContentMine/go-europmc
* https://github.com/jlaffaye/ftp
