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
